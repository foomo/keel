package keel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/markdown"
	"github.com/foomo/keel/service"
	"github.com/foomo/keel/telemetry/nonrecording"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
)

// Server struct
type Server struct {
	services        []Service
	initServices    []Service
	meter           metric.Meter
	meterProvider   metric.MeterProvider
	tracer          trace.Tracer
	traceProvider   trace.TracerProvider
	shutdownSignals []os.Signal
	shutdownTimeout time.Duration
	running         atomic.Bool
	closers         []interface{}
	closersLock     sync.Mutex
	probes          map[healthz.Type][]interface{}
	documenter      map[string]interfaces.Documenter
	ctx             context.Context
	ctxCancel       context.Context
	ctxCancelFn     context.CancelFunc
	g               *errgroup.Group
	gCtx            context.Context
	l               *zap.Logger
	c               *viper.Viper
}

func NewServer(opts ...Option) *Server {
	inst := &Server{
		shutdownTimeout: 30 * time.Second,
		shutdownSignals: []os.Signal{os.Interrupt, syscall.SIGTERM},
		probes:          map[healthz.Type][]interface{}{},
		documenter:      map[string]interfaces.Documenter{},
		ctx:             context.Background(),
		c:               config.Config(),
		l:               log.Logger(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	{ // setup error group
		inst.ctxCancel, inst.ctxCancelFn = signal.NotifyContext(inst.ctx, inst.shutdownSignals...)
		inst.g, inst.gCtx = errgroup.WithContext(inst.ctxCancel)

		// gracefully shutdown
		inst.g.Go(func() error {
			<-inst.gCtx.Done()
			inst.l.Debug("keel graceful shutdown")
			defer inst.ctxCancelFn()

			timeoutCtx, timeoutCancel := context.WithTimeout(inst.ctx, inst.shutdownTimeout)
			defer timeoutCancel()

			// append internal closers
			inst.closersLock.Lock()
			defer inst.closersLock.Unlock()
			closers := append(inst.closers, inst.traceProvider, inst.meterProvider) //nolint:gocritic

			for _, closer := range closers {
				l := inst.l.With(log.FName(fmt.Sprintf("%T", closer)))
				switch c := closer.(type) {
				case interfaces.Closer:
					c.Close()
				case interfaces.ErrorCloser:
					if err := c.Close(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorCloser")
					}
				case interfaces.CloserWithContext:
					c.Close(timeoutCtx)
				case interfaces.ErrorCloserWithContext:
					if err := c.Close(timeoutCtx); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorCloserWithContext")
					}
				case interfaces.Shutdowner:
					c.Shutdown()
				case interfaces.ErrorShutdowner:
					if err := c.Shutdown(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorShutdowner")
					}
				case interfaces.ShutdownerWithContext:
					c.Shutdown(timeoutCtx)
				case interfaces.ErrorShutdownerWithContext:
					if err := c.Shutdown(timeoutCtx); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorShutdownerWithContext")
					}
				case interfaces.Stopper:
					c.Stop()
				case interfaces.ErrorStopper:
					if err := c.Stop(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorStopper")
					}
				case interfaces.StopperWithContext:
					c.Stop(timeoutCtx)
				case interfaces.ErrorStopperWithContext:
					if err := c.Stop(timeoutCtx); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorStopperWithContext")
					}
				case interfaces.Unsubscriber:
					c.Unsubscribe()
				case interfaces.ErrorUnsubscriber:
					if err := c.Unsubscribe(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorUnsubscriber")
					}
				case interfaces.UnsubscriberWithContext:
					c.Unsubscribe(timeoutCtx)
				case interfaces.ErrorUnsubscriberWithContext:
					if err := c.Unsubscribe(timeoutCtx); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorUnsubscriberWithContext")
					}
				}
			}
			return inst.gCtx.Err()
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
	inst.AddDocumenter("Keel Server", inst)

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
	return s.ctxCancel
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	for _, value := range s.services {
		if value == service {
			return
		}
	}
	s.services = append(s.services, service)
	s.AddAlwaysHealthzers(service)
	s.AddCloser(service)
}

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddCloser adds a closer to be called on shutdown
func (s *Server) AddCloser(closer interface{}) {
	s.closersLock.Lock()
	defer s.closersLock.Unlock()
	for _, value := range s.closers {
		if value == closer {
			return
		}
	}
	switch closer.(type) {
	case interfaces.Closer,
		interfaces.ErrorCloser,
		interfaces.CloserWithContext,
		interfaces.ErrorCloserWithContext,
		interfaces.Shutdowner,
		interfaces.ErrorShutdowner,
		interfaces.ShutdownerWithContext,
		interfaces.ErrorShutdownerWithContext,
		interfaces.Stopper,
		interfaces.ErrorStopper,
		interfaces.StopperWithContext,
		interfaces.ErrorStopperWithContext,
		interfaces.Unsubscriber,
		interfaces.ErrorUnsubscriber,
		interfaces.UnsubscriberWithContext,
		interfaces.ErrorUnsubscriberWithContext:
		s.closers = append(s.closers, closer)
	default:
		s.l.Warn("unable to add closer", log.FValue(fmt.Sprintf("%T", closer)))
	}
}

// AddClosers adds the given closers to be called on shutdown
func (s *Server) AddClosers(closers ...interface{}) {
	for _, closer := range closers {
		s.AddCloser(closer)
	}
}

// AddDocumenter adds a dcoumenter to beadded to the exposed docs
func (s *Server) AddDocumenter(name string, documenter interfaces.Documenter) {
	s.documenter[name] = documenter
}

// AddHealthzer adds a probe to be called on healthz checks
func (s *Server) AddHealthzer(typ healthz.Type, probe interface{}) {
	switch probe.(type) {
	case healthz.BoolHealthzer,
		healthz.BoolHealthzerWithContext,
		healthz.ErrorHealthzer,
		healthz.ErrorHealthzWithContext,
		interfaces.ErrorPinger,
		interfaces.ErrorPingerWithContext:
		s.probes[typ] = append(s.probes[typ], probe)
	default:
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
	return errors.Is(s.gCtx.Err(), context.Canceled)
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

	defer s.ctxCancelFn()
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

// Docs returns the self-documenting string
func (s *Server) Docs() string {
	md := &markdown.Markdown{}

	{
		var rows [][]string
		keys := s.Config().AllKeys()
		defaults := config.Defaults()
		for _, key := range keys {
			var fallback interface{}
			if v, ok := defaults[key]; ok {
				fallback = v
			}
			rows = append(rows, []string{
				markdown.Code(key),
				markdown.Code(config.TypeOf(key)),
				"",
				markdown.Code(fmt.Sprintf("%v", fallback)),
			})
		}
		for _, key := range config.RequiredKeys() {
			rows = append(rows, []string{
				markdown.Code(key),
				markdown.Code(config.TypeOf(key)),
				markdown.Code("true"),
				"",
			})
		}
		if len(rows) > 0 {
			md.Println("### Config")
			md.Println("")
			md.Println("List of all registered config variabled with their defaults.")
			md.Println("")
			md.Table([]string{"Key", "Type", "Required", "Default"}, rows)
			md.Println("")
		}
	}

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
			md.Println("")
		}
	}

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
			md.Println("### Services")
			md.Println("")
			md.Println("List of all registered services that are being started.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Description"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		for k, probes := range s.probes {
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
			md.Println("List of all registered healthz probes that are being called during startup and runntime.")
			md.Println("")
			md.Table([]string{"Name", "Probe", "Type", "Description"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		s.closersLock.Lock()
		defer s.closersLock.Unlock()
		for _, value := range s.closers {
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
	}

	{
		var rows [][]string
		s.meter.AsyncFloat64()

		values := nonrecording.Metrics()

		gatherer, _ := prometheus.DefaultRegisterer.(*prometheus.Registry).Gather()
		for _, value := range gatherer {
			values = append(values, nonrecording.Metric{
				Name: value.GetName(),
				Type: value.GetType().String(),
				Help: value.GetHelp(),
			})
		}
		for _, value := range values {
			rows = append(rows, []string{
				markdown.Code(value.Name),
				value.Type,
				value.Help,
			})
		}
		if len(rows) > 0 {
			md.Println("### Metrics")
			md.Println("")
			md.Println("List of all registered metrics than are being exposed.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Description"}, rows)
			md.Println("")
		}
	}

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
