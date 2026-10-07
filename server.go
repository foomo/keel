package keel

import (
	"context"
	"errors"
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
	internalotel "github.com/foomo/keel/internal/otel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/markdown"
	keelhttp "github.com/foomo/keel/net/http"
	keelsemconv "github.com/foomo/keel/semconv"
	"github.com/foomo/keel/service"
	"github.com/foomo/keel/telemetry"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

// Names and default addresses of the HTTP services registered by
// [Server.AddInternalHTTPService] and [Server.AddPublicHTTPService].
const (
	ServiceNameInternalHTTP = "internal"
	ServiceAddrInternalHTTP = ":8080"
	ServiceNamePublicHTTP   = "public"
	ServiceAddrPublicHTTP   = ":8000"
)

// Server runs a set of [Service] values until a shutdown signal is received and
// then shuts them down gracefully. Create it with [NewServer]; the zero value is
// not usable. Registration methods are safe for concurrent use.
type Server struct {
	services        []Service
	initServices    []Service
	meterProvider   metric.MeterProvider
	traceProvider   trace.TracerProvider
	loggerProvider  otellog.LoggerProvider
	shutdown        atomic.Bool
	shutdownSignals []os.Signal
	// gracefulPeriod should equal the terminationGracePeriodSeconds
	gracefulPeriod   time.Duration
	running          atomic.Bool
	syncClosers      []any
	syncClosersLock  sync.RWMutex
	syncReadmers     []interfaces.Readmer
	syncReadmersLock sync.RWMutex
	syncProbes       map[healthz.Type][]any
	syncProbesLock   sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	gracefulCtx      context.Context
	gracefulCancel   context.CancelFunc
	g                *errgroup.Group
	gCtx             context.Context
	l                *zap.Logger
	c                *viper.Viper
}

// NewServer creates a [Server] configured by opts and immediately starts its
// init services. It sets the global OpenTelemetry logger, error handler and
// text map propagator, falls back to no-op trace, meter and logger providers
// when none are configured, and registers the server itself as an always
// probe and a readiness probe that fails once shutdown has started.
func NewServer(opts ...Option) *Server {
	inst := &Server{
		gracefulPeriod:  time.Duration(env.GetInt("KEEL_GRACEFUL_PERIOD", 30)) * time.Second,
		shutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		syncReadmers:    []interfaces.Readmer{},
		syncProbes:      map[healthz.Type][]any{},
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

		inst.ctx, inst.cancel = context.WithCancel(inst.ctx)
		inst.g, inst.gCtx = errgroup.WithContext(inst.ctx)
		inst.gracefulCtx, inst.gracefulCancel = signal.NotifyContext(inst.gCtx, inst.shutdownSignals...)

		// gracefully shutdown
		inst.g.Go(func() error {
			<-inst.gracefulCtx.Done()
			inst.shutdown.Store(true)

			// detach from the server context so closers get the full graceful period
			// even if the server context has been canceled
			timeoutCtx, timeoutCancel := context.WithTimeout(context.WithoutCancel(inst.ctx), inst.gracefulPeriod)
			defer timeoutCancel()

			inst.l.Info("keel closer closed",
				log.Attribute(keelsemconv.KeelGracefulPeriod(inst.gracefulPeriod)),
			)

			// append internal closers
			closers := append(inst.closers(), inst.traceProvider, inst.meterProvider, inst.loggerProvider)

			inst.l.Info("keel closer closed: closers")

			closeAll(timeoutCtx, inst.l, closers)

			inst.l.Info("keel closer closed: complete")

			return ErrServerShutdown
		})
	}

	{ // setup telemetry
		otel.SetLogger(logr.New(internalotel.NewLogger(inst.l)))
		otel.SetErrorHandler(internalotel.NewErrorHandler(inst.l))
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		if inst.meterProvider == nil {
			inst.meterProvider = telemetry.NewNoopMeterProvider()
		}

		if inst.traceProvider == nil {
			inst.traceProvider = telemetry.NewNoopTraceProvider()
		}

		if inst.loggerProvider == nil {
			inst.loggerProvider = telemetry.NewNoopLoggerProvider()
		} else {
			inst.l = inst.l.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewTee(c, telemetry.NewZapBridgeCore(inst.loggerProvider))
			}))
		}
	}

	// add probe
	inst.AddAlwaysHealthzers(inst)
	//nolint:staticcheck //SA1019
	inst.AddReadmers(
		interfaces.ReadmeFunc(env.Readme),
		interfaces.ReadmeFunc(config.Readme),
		inst,
		interfaces.ReadmeFunc(telemetry.Readme),
	)

	// start init services
	inst.startService(inst.initServices...)

	return inst
}

// Logger returns the server logger.
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Meter returns the global OpenTelemetry meter from [telemetry.Meter].
func (s *Server) Meter() metric.Meter {
	return telemetry.Meter()
}

// Tracer returns the global OpenTelemetry tracer from [telemetry.Tracer].
func (s *Server) Tracer() trace.Tracer {
	return telemetry.Tracer()
}

// Config returns the server configuration.
func (s *Server) Config() *viper.Viper {
	return s.c
}

// Context returns the server context. It is canceled when [Server.Run] returns.
func (s *Server) Context() context.Context {
	return s.ctx
}

// ShutdownContext returns a context that is canceled when graceful shutdown is
// triggered by a shutdown signal, [Server.ShutdownCancel] or a failing service.
func (s *Server) ShutdownContext() context.Context {
	return s.gracefulCtx
}

// ShutdownCancel returns a function that triggers graceful shutdown.
func (s *Server) ShutdownCancel() context.CancelFunc {
	return s.gracefulCancel
}

// AddService registers v to be started by [Server.Run], as an always probe and
// as a closer. Adding the same service twice is a no-op.
func (s *Server) AddService(v Service) {
	if !slices.Contains(s.services, v) {
		s.services = append(s.services, v)
		s.AddAlwaysHealthzers(v)
		s.AddCloser(v)
	}
}

// AddHTTPService registers a [service.HTTP] named name serving handler wrapped
// in middleware. The listen address is read from the config key
// service.http.<name>.addr and defaults to addr.
func (s *Server) AddHTTPService(name, addr string, handler http.Handler, middleware ...keelhttp.Middleware) {
	addrFn := config.GetString(s.Config(), "service.http."+name+".addr", addr)
	s.AddService(service.NewHTTP(s.l, name, addrFn(), handler, middleware...))
}

// AddInternalHTTPService registers an HTTP service named
// [ServiceNameInternalHTTP] listening on [ServiceAddrInternalHTTP] by default.
// See [Server.AddHTTPService].
func (s *Server) AddInternalHTTPService(handler http.Handler, middleware ...keelhttp.Middleware) {
	s.AddHTTPService(ServiceNameInternalHTTP, ServiceAddrInternalHTTP, handler, middleware...)
}

// AddPublicHTTPService registers an HTTP service named [ServiceNamePublicHTTP]
// listening on [ServiceAddrPublicHTTP] by default. See [Server.AddHTTPService].
func (s *Server) AddPublicHTTPService(handler http.Handler, middleware ...keelhttp.Middleware) {
	s.AddHTTPService(ServiceNamePublicHTTP, ServiceAddrPublicHTTP, handler, middleware...)
}

// AddGoRoutine registers a [service.GoRoutine] named name running handler.
func (s *Server) AddGoRoutine(name string, handler service.GoRoutineFn, opts ...service.GoRoutineOption) {
	s.AddService(service.NewGoRoutine(s.l, name, handler, opts...))
}

// AddServices calls [Server.AddService] for each of services.
func (s *Server) AddServices(services ...Service) {
	for _, value := range services {
		s.AddService(value)
	}
}

// AddCloser registers closer to be called on graceful shutdown. Closers are
// called in registration order. A warning is logged if closer does not satisfy
// [IsCloser]; adding the same closer twice is a no-op.
func (s *Server) AddCloser(closer any) {
	if !IsCloser(closer) {
		s.l.Warn("unable to add closer", log.Attribute(keelsemconv.KeelCloserType(closer)))
	}

	if slices.Contains(s.closers(), closer) {
		return
	}

	s.addClosers(closer)
}

// AddClosers calls [Server.AddCloser] for each of closers.
func (s *Server) AddClosers(closers ...any) {
	for _, closer := range closers {
		s.AddCloser(closer)
	}
}

// AddReadmer adds readmer to the readme exposed by the readme service.
//
// Deprecated: Readme support will be removed in a future release.
func (s *Server) AddReadmer(readmer interfaces.Readmer) {
	s.addReadmers(readmer)
}

// AddReadmers adds readmers to the readme exposed by the readme service.
//
// Deprecated: Readme support will be removed in a future release.
func (s *Server) AddReadmers(readmers ...interfaces.Readmer) {
	for _, readmer := range readmers {
		s.AddReadmer(readmer)
	}
}

// AddHealthzer registers probe for health checks of type typ. Values that do not
// satisfy [IsHealthz] are ignored.
func (s *Server) AddHealthzer(typ healthz.Type, probe any) {
	if IsHealthz(probe) {
		s.addProbes(typ, probe)
	} else {
		s.l.Debug("not a healthz probe", log.Attribute(keelsemconv.KeelHealthzProbeType(probe)))
	}
}

// AddHealthzers calls [Server.AddHealthzer] for each of probes.
func (s *Server) AddHealthzers(typ healthz.Type, probes ...any) {
	for _, probe := range probes {
		s.AddHealthzer(typ, probe)
	}
}

// AddAlwaysHealthzers registers probes that are checked by every probe type.
func (s *Server) AddAlwaysHealthzers(probes ...any) {
	s.AddHealthzers(healthz.TypeAlways, probes...)
}

// AddStartupHealthzers registers probes for the startup check.
func (s *Server) AddStartupHealthzers(probes ...any) {
	s.AddHealthzers(healthz.TypeStartup, probes...)
}

// AddLivenessHealthzers registers probes for the liveness check.
func (s *Server) AddLivenessHealthzers(probes ...any) {
	s.AddHealthzers(healthz.TypeLiveness, probes...)
}

// AddReadinessHealthzers registers probes for the readiness check.
func (s *Server) AddReadinessHealthzers(probes ...any) {
	s.AddHealthzers(healthz.TypeReadiness, probes...)
}

// Healthz returns [ErrServerNotRunning] unless [Server.Run] is in progress.
func (s *Server) Healthz() error {
	if !s.running.Load() {
		return ErrServerNotRunning
	}

	return nil
}

// Readme returns a markdown description of the registered services, health
// probes and closers.
func (s *Server) Readme() string {
	md := &markdown.Markdown{}

	md.Println(s.readmeServices())
	md.Println(s.readmeHealthz())
	md.Print(s.readmeCloser())

	return md.String()
}

// Run starts the registered services and blocks until the server has shut down,
// either gracefully after a shutdown signal or because a service failed. The
// outcome is logged; Run does not exit the process.
func (s *Server) Run() {
	s.l.With(log.Attributes(telemetry.EnvAttributes()...)...).Info("starting keel server")
	defer s.cancel()

	// start services
	s.startService(s.services...)

	// add init services to closers
	for _, initService := range s.initServices {
		s.AddClosers(initService)
	}

	// set running
	defer s.running.Store(false)

	s.running.Store(true)

	// wait for shutdown
	if err := s.g.Wait(); errors.Is(err, ErrServerShutdown) {
		s.l.Info("keel server stopped")
	} else if err != nil {
		log.WithError(s.l, err).Error("keel server failed")
	}
}

func (s *Server) closers() []any {
	s.syncClosersLock.RLock()
	defer s.syncClosersLock.RUnlock()

	return s.syncClosers
}

func (s *Server) addClosers(v ...any) {
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

func (s *Server) probes() map[healthz.Type][]any {
	s.syncProbesLock.RLock()
	defer s.syncProbesLock.RUnlock()

	return s.syncProbes
}

func (s *Server) addProbes(typ healthz.Type, v ...any) {
	s.syncProbesLock.Lock()
	defer s.syncProbesLock.Unlock()

	s.syncProbes[typ] = append(s.syncProbes[typ], v...)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

// startService starts each of services in the server's error group with the
// server context. A service returning [net/http.ErrServerClosed] is treated as
// stopped; any other error fails the group and triggers shutdown.
func (s *Server) startService(services ...Service) {
	done := make(chan struct{}, 1)

	for _, value := range services {
		s.g.Go(func() error {
			done <- struct{}{}

			if err := value.Start(s.ctx); errors.Is(err, http.ErrServerClosed) {
				log.WithError(s.l, err).Debug("server has closed")
			} else if err != nil {
				log.WithError(s.l, err).Error("failed to start service")

				return err
			}

			return nil
		})
		<-done
	}

	close(done)
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
				t := reflect.TypeFor[*service.HTTP]()
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
				markdown.Code(t.Name()),
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
