package keel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/markdown"
	"github.com/foomo/keel/metrics"
	"github.com/foomo/keel/service"
	"github.com/foomo/keel/telemetry"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Server struct
type Server struct {
	services        []Service
	initServices    []Service
	meter           metric.Meter
	meterProvider   metric.MeterProvider
	tracer          trace.Tracer
	traceProvider   trace.TracerProvider
	shutdown        atomic.Bool
	shutdownSignals []os.Signal
	// gracefulTimeout should equal the readinessProbe's periodSeconds * failureThreshold
	gracefulTimeout time.Duration
	// shutdownTimeout should equal the readinessProbe's terminationGracePeriodSeconds
	shutdownTimeout  time.Duration
	running          atomic.Bool
	syncClosers      []interface{}
	syncClosersLock  sync.RWMutex
	syncReadmers     []interfaces.Readmer
	syncReadmersLock sync.RWMutex
	syncProbes       map[healthz.Type][]interface{}
	syncProbesLock   sync.RWMutex
	ctx              context.Context
	cancelCtx        context.Context
	cancelFunc       context.CancelFunc
	shutdownCtx      context.Context
	shutdownFunc     context.CancelFunc
	g                *errgroup.Group
	gCtx             context.Context
	l                *zap.Logger
	c                *viper.Viper
}

func NewServer(opts ...Option) *Server {
	inst := &Server{
		gracefulTimeout: time.Duration(env.GetInt("KEEL_GRACEFUL_TIMEOUT", 10*3)) * time.Second,
		shutdownTimeout: time.Duration(env.GetInt("KEEL_SHUTDOWN_TIMEOUT", 30)) * time.Second,
		shutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		syncReadmers:    []interfaces.Readmer{},
		syncProbes:      map[healthz.Type][]interface{}{},
		ctx:             context.Background(),
		c:               config.Config(),
		l:               log.Logger(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	{ // setup error group
		inst.AddReadinessHealthzers(healthz.NewHealthzerFn(func(ctx context.Context) error {
			if inst.shutdown.Load() {
				return ErrServerShutdown
			}
			return nil
		}))

		inst.cancelCtx, inst.cancelFunc = context.WithCancel(inst.ctx)
		inst.g, inst.gCtx = errgroup.WithContext(inst.cancelCtx)
		inst.shutdownCtx, inst.shutdownFunc = signal.NotifyContext(inst.ctx, inst.shutdownSignals...)

		// gracefully shutdown
		inst.g.Go(func() error {
			<-inst.shutdownCtx.Done()
			inst.l.Info("keel graceful shutdown")

			timeoutCtx, timeoutCancel := context.WithTimeout(inst.ctx, inst.shutdownTimeout)
			defer timeoutCancel()

			inst.shutdown.Store(true)

			inst.l.Info("keel graceful shutdown timeout",
				zap.Duration("graceful_timeout", inst.gracefulTimeout),
				zap.Duration("shutdown_timeout", inst.shutdownTimeout),
			)
			{
				timer := time.NewTimer(inst.gracefulTimeout)
				select {
				case <-timeoutCtx.Done():
					timer.Stop()
				case <-timer.C:
				}
			}
			inst.l.Info("keel graceful shutdown timeout complete")

			// append internal closers
			closers := append(inst.closers(), inst.traceProvider, inst.meterProvider)

			inst.l.Info("keel graceful shutdown closers")
			for _, closer := range closers {
				var err error
				l := inst.l.With(log.FName(fmt.Sprintf("%T", closer)))
				switch c := closer.(type) {
				case interfaces.Closer:
					c.Close()
				case interfaces.ErrorCloser:
					err = c.Close()
				case interfaces.CloserWithContext:
					c.Close(timeoutCtx)
				case interfaces.ErrorCloserWithContext:
					err = c.Close(timeoutCtx)
				case interfaces.Shutdowner:
					c.Shutdown()
				case interfaces.ErrorShutdowner:
					err = c.Shutdown()
				case interfaces.ShutdownerWithContext:
					c.Shutdown(timeoutCtx)
				case interfaces.ErrorShutdownerWithContext:
					err = c.Shutdown(timeoutCtx)
				case interfaces.Stopper:
					c.Stop()
				case interfaces.ErrorStopper:
					err = c.Stop()
				case interfaces.StopperWithContext:
					c.Stop(timeoutCtx)
				case interfaces.ErrorStopperWithContext:
					err = c.Stop(timeoutCtx)
				case interfaces.Unsubscriber:
					c.Unsubscribe()
				case interfaces.ErrorUnsubscriber:
					err = c.Unsubscribe()
				case interfaces.UnsubscriberWithContext:
					c.Unsubscribe(timeoutCtx)
				case interfaces.ErrorUnsubscriberWithContext:
					err = c.Unsubscribe(timeoutCtx)
				}
				if err != nil {
					l.Warn("keel graceful shutdown closer failed", zap.Error(err))
				} else {
					l.Debug("keel graceful shutdown closer closed")
				}
			}

			inst.l.Info("keel graceful shutdown complete")

			return nil
		})
	}

	{ // setup telemetry
		var err error
		otel.SetLogger(logr.New(telemetry.NewLogger(inst.l)))
		otel.SetErrorHandler(telemetry.NewErrorHandler(inst.l))
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		if inst.meterProvider == nil {
			inst.meterProvider, err = telemetry.NewNoopMeterProvider()
			log.Must(inst.l, err, "failed to create meter provider")
		} else if env.GetBool("OTEL_ENABLED", false) {
			if env.GetBool("OTEL_METRICS_HOST_ENABLED", false) {
				log.Must(inst.l, otelhost.Start(), "failed to start otel host metrics")
			}
			if env.GetBool("OTEL_METRICS_RUNTIME_ENABLED", false) {
				log.Must(inst.l, otelruntime.Start(), "failed to start otel runtime metrics")
			}
		}
		inst.meter = telemetry.Meter()

		if inst.traceProvider == nil {
			inst.traceProvider, err = telemetry.NewNoopTraceProvider()
			log.Must(inst.l, err, "failed to create tracer provider")
		}
		inst.tracer = telemetry.Tracer()
	}

	// add probe
	inst.AddAlwaysHealthzers(inst)
	inst.AddReadmers(
		interfaces.ReadmeFunc(env.Readme),
		interfaces.ReadmeFunc(config.Readme),
		inst,
		interfaces.ReadmeFunc(metrics.Readme),
	)

	// start init services
	inst.startService(inst.initServices...)

	return inst
}

// Logger returns server logger
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Meter returns the implementation meter
func (s *Server) Meter() metric.Meter {
	return s.meter
}

// Tracer returns the implementation tracer
func (s *Server) Tracer() trace.Tracer {
	return s.tracer
}

// Config returns server config
func (s *Server) Config() *viper.Viper {
	return s.c
}

// Context returns server context
func (s *Server) Context() context.Context {
	return s.ctx
}

// CancelContext returns server's cancel context
func (s *Server) CancelContext() context.Context {
	return s.cancelCtx
}

// CancelFunc returns server's cancel function
func (s *Server) CancelFunc() context.CancelFunc {
	return s.cancelFunc
}

// ShutdownContext returns server's shutdown cancel context
func (s *Server) ShutdownContext() context.Context {
	return s.shutdownCtx
}

// ShutdownCancel returns server's shutdown cancel function
func (s *Server) ShutdownCancel() context.CancelFunc {
	return s.shutdownFunc
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	if !slices.Contains(s.services, service) {
		s.services = append(s.services, service)
		s.AddAlwaysHealthzers(service)
		s.AddCloser(service)
	}
}

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, value := range services {
		s.AddService(value)
	}
}

// AddCloser adds a closer to be called on shutdown
func (s *Server) AddCloser(closer interface{}) {
	for _, value := range s.closers() {
		if value == closer {
			return
		}
	}
	if IsCloser(closer) {
		s.addClosers(closer)
	} else {
		s.l.Warn("unable to add closer", log.FValue(fmt.Sprintf("%T", closer)))
	}
}

// AddClosers adds the given closers to be called on shutdown
func (s *Server) AddClosers(closers ...interface{}) {
	for _, closer := range closers {
		s.AddCloser(closer)
	}
}

// AddReadmer adds a readmer to be added to the exposed readme
func (s *Server) AddReadmer(readmer interfaces.Readmer) {
	s.addReadmers(readmer)
}

// AddReadmers adds readmers to be added to the exposed readme
func (s *Server) AddReadmers(readmers ...interfaces.Readmer) {
	for _, readmer := range readmers {
		s.AddReadmer(readmer)
	}
}

// AddHealthzer adds a probe to be called on healthz checks
func (s *Server) AddHealthzer(typ healthz.Type, probe interface{}) {
	if IsHealthz(probe) {
		s.addProbes(typ, probe)
	} else {
		s.l.Debug("not a healthz probe", log.FValue(fmt.Sprintf("%T", probe)))
	}
}

// AddHealthzers adds the given probes to be called on healthz checks
func (s *Server) AddHealthzers(typ healthz.Type, probes ...interface{}) {
	for _, probe := range probes {
		s.AddHealthzer(typ, probe)
	}
}

// AddAlwaysHealthzers adds the probes to be called on any healthz checks
func (s *Server) AddAlwaysHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeAlways, probes...)
}

// AddStartupHealthzers adds the startup probes to be called on healthz checks
func (s *Server) AddStartupHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeStartup, probes...)
}

// AddLivenessHealthzers adds the liveness probes to be called on healthz checks
func (s *Server) AddLivenessHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeLiveness, probes...)
}

// AddReadinessHealthzers adds the readiness probes to be called on healthz checks
func (s *Server) AddReadinessHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeReadiness, probes...)
}

// IsCanceled returns true if the internal errgroup has been canceled
func (s *Server) IsCanceled() bool {
	return s.cancelCtx.Err() != nil
}

// Healthz returns true if the server is running
func (s *Server) Healthz() error {
	if !s.running.Load() {
		return ErrServerNotRunning
	}
	return nil
}

// Run runs the server
func (s *Server) Run() {
	if s.IsCanceled() {
		s.l.Info("keel server canceled")
		return
	}

	defer s.cancelFunc()
	s.l.Info("starting keel server")

	// start services
	s.startService(s.services...)

	// add init services to closers
	for _, initService := range s.initServices {
		s.AddClosers(initService)
	}

	// set running
	defer func() {
		s.running.Store(false)
	}()
	s.running.Store(true)

	// wait for shutdown
	if err := s.g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.WithError(s.l, err).Error("service error")
	}

	s.l.Info("keel server stopped")
}

func (s *Server) closers() []interface{} {
	s.syncClosersLock.RLock()
	defer s.syncClosersLock.RUnlock()
	return s.syncClosers
}

func (s *Server) addClosers(v ...interface{}) {
	s.syncClosersLock.Lock()
	defer s.syncClosersLock.Unlock()
	s.syncClosers = append(s.syncClosers, v...)
}

func (s *Server) readmers() []interfaces.Readmer {
	s.syncReadmersLock.RLock()
	defer s.syncReadmersLock.RUnlock()
	return s.syncReadmers
}

func (s *Server) addReadmers(v ...interfaces.Readmer) {
	s.syncReadmersLock.Lock()
	defer s.syncReadmersLock.Unlock()
	s.syncReadmers = append(s.syncReadmers, v...)
}

func (s *Server) probes() map[healthz.Type][]interface{} {
	s.syncProbesLock.RLock()
	defer s.syncProbesLock.RUnlock()
	return s.syncProbes
}

func (s *Server) addProbes(typ healthz.Type, v ...interface{}) {
	s.syncProbesLock.Lock()
	defer s.syncProbesLock.Unlock()
	s.syncProbes[typ] = append(s.syncProbes[typ], v...)
}

// Readme returns the self-documenting string
func (s *Server) Readme() string {
	md := &markdown.Markdown{}

	md.Println(s.readmeServices())
	md.Println(s.readmeHealthz())
	md.Print(s.readmeCloser())

	return md.String()
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

// startService starts the given services
func (s *Server) startService(services ...Service) {
	for _, value := range services {
		value := value
		s.g.Go(func() error {
			if err := value.Start(s.ctx); errors.Is(err, http.ErrServerClosed) {
				log.WithError(s.l, err).Debug("server has closed")
			} else if err != nil {
				log.WithError(s.l, err).Error("failed to start service")
				return err
			}
			return nil
		})
	}
}

func (s *Server) readmeCloser() string {
	md := &markdown.Markdown{}
	closers := s.closers()
	rows := make([][]string, 0, len(closers))
	for _, value := range closers {
		t := reflect.TypeOf(value)
		var closer string
		switch value.(type) {
		case interfaces.Closer:
			closer = "Closer"
		case interfaces.ErrorCloser:
			closer = "ErrorCloser"
		case interfaces.CloserWithContext:
			closer = "CloserWithContext"
		case interfaces.ErrorCloserWithContext:
			closer = "ErrorCloserWithContext"
		case interfaces.Shutdowner:
			closer = "Shutdowner"
		case interfaces.ErrorShutdowner:
			closer = "ErrorShutdowner"
		case interfaces.ShutdownerWithContext:
			closer = "ShutdownerWithContext"
		case interfaces.ErrorShutdownerWithContext:
			closer = "ErrorShutdownerWithContext"
		case interfaces.Stopper:
			closer = "Stopper"
		case interfaces.ErrorStopper:
			closer = "ErrorStopper"
		case interfaces.StopperWithContext:
			closer = "StopperWithContext"
		case interfaces.ErrorStopperWithContext:
			closer = "ErrorStopperWithContext"
		case interfaces.Unsubscriber:
			closer = "Unsubscriber"
		case interfaces.ErrorUnsubscriber:
			closer = "ErrorUnsubscriber"
		case interfaces.UnsubscriberWithContext:
			closer = "UnsubscriberWithContext"
		case interfaces.ErrorUnsubscriberWithContext:
			closer = "ErrorUnsubscriberWithContext"
		}
		rows = append(rows, []string{
			markdown.Code(markdown.Name(value)),
			markdown.Code(t.String()),
			markdown.Code(closer),
			markdown.String(value),
		})
	}
	if len(rows) > 0 {
		md.Println("### Closers")
		md.Println("")
		md.Println("List of all registered closers that are being called during graceful shutdown.")
		md.Println("")
		md.Table([]string{"Name", "Type", "Closer", "Description"}, rows)
		md.Println("")
	}

	return md.String()
}

func (s *Server) readmeHealthz() string {
	var rows [][]string
	md := &markdown.Markdown{}

	for k, probes := range s.probes() {
		for _, probe := range probes {
			t := reflect.TypeOf(probe)
			rows = append(rows, []string{
				markdown.Code(markdown.Name(probe)),
				markdown.Code(k.String()),
				markdown.Code(t.String()),
				markdown.String(probe),
			})
		}
	}
	if len(rows) > 0 {
		md.Println("### Health probes")
		md.Println("")
		md.Println("List of all registered healthz probes that are being called during startup and runtime.")
		md.Println("")
		md.Table([]string{"Name", "Probe", "Type", "Description"}, rows)
	}

	return md.String()
}

func (s *Server) readmeServices() string {
	md := &markdown.Markdown{}

	{
		var rows [][]string
		for _, value := range s.initServices {
			if v, ok := value.(*service.HTTP); ok {
				t := reflect.TypeOf(v)
				rows = append(rows, []string{
					markdown.Code(v.Name()),
					markdown.Code(t.String()),
					markdown.String(v),
				})
			}
		}
		if len(rows) > 0 {
			md.Println("### Init Services")
			md.Println("")
			md.Println("List of all registered init services that are being immediately started.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Address"}, rows)
		}
	}

	md.Println("")

	{
		var rows [][]string
		for _, value := range s.services {
			t := reflect.TypeOf(value)
			rows = append(rows, []string{
				markdown.Code(value.Name()),
				markdown.Code(t.String()),
				markdown.String(value),
			})
		}
		if len(rows) > 0 {
			md.Println("### Runtime Services")
			md.Println("")
			md.Println("List of all registered services that are being started.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Description"}, rows)
		}
	}

	return md.String()
}
