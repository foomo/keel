package keel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
	meter           metric.MeterMust
	meterProvider   metric.MeterProvider
	tracer          trace.Tracer
	traceProvider   trace.TracerProvider
	shutdownSignals []os.Signal
	shutdownTimeout time.Duration
	running         bool
	closers         []interface{}
	probes          map[HealthzType][]interface{}
	ctx             context.Context
	ctxCancel       context.CancelFunc
	g               *errgroup.Group
	gCtx            context.Context
	l               *zap.Logger
	c               *viper.Viper
}

func NewServer(opts ...Option) *Server {
	inst := &Server{
		shutdownTimeout: 30 * time.Second,
		shutdownSignals: []os.Signal{os.Interrupt, syscall.SIGTERM},
		probes:          map[HealthzType][]interface{}{},
		ctx:             context.Background(),
		c:               config.Config(),
		l:               log.Logger(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	{ // setup error group
		var ctx context.Context
		ctx, inst.ctxCancel = signal.NotifyContext(inst.ctx, inst.shutdownSignals...)
		inst.g, inst.gCtx = errgroup.WithContext(ctx)

		// gracefully shutdown
		inst.g.Go(func() error {
			<-inst.gCtx.Done()
			inst.l.Debug("keel graceful shutdown")
			defer inst.ctxCancel()

			timeoutCtx, timeoutCancel := context.WithTimeout(inst.ctx, inst.shutdownTimeout)
			defer timeoutCancel()

			// append internal closers
			closers := append(inst.closers, inst.traceProvider, inst.meterProvider) //nolint:gocritic

			for _, closer := range closers {
				l := inst.l.With(log.FName(fmt.Sprintf("%T", closer)))
				switch c := closer.(type) {
				case Closer:
					c.Close()
				case ErrorCloser:
					if err := c.Close(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorCloser")
					}
				case CloserWithContext:
					c.Close(timeoutCtx)
				case ErrorCloserWithContext:
					if err := c.Close(timeoutCtx); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorCloserWithContext")
					}
				case Shutdowner:
					c.Shutdown()
				case ErrorShutdowner:
					if err := c.Shutdown(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorShutdowner")
					}
				case ShutdownerWithContext:
					c.Shutdown(timeoutCtx)
				case ErrorShutdownerWithContext:
					if err := c.Shutdown(timeoutCtx); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorShutdownerWithContext")
					}
				case Unsubscriber:
					c.Unsubscribe()
				case ErrorUnsubscriber:
					if err := c.Unsubscribe(); err != nil {
						log.WithError(l, err).Error("failed to gracefully stop ErrorUnsubscriber")
					}
				case UnsubscriberWithContext:
					c.Unsubscribe(timeoutCtx)
				case ErrorUnsubscriberWithContext:
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
		inst.meter = telemetry.MustMeter()

		if inst.traceProvider == nil {
			inst.traceProvider, err = telemetry.NewNoopTraceProvider()
			log.Must(inst.l, err, "failed to create tracer provider")
		}
		inst.tracer = telemetry.Tracer()
	}

	// add probe
	inst.AddAlwaysHealthzers(inst)

	// start init services
	inst.startService(inst.initServices...)

	return inst
}

// Logger returns server logger
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Meter returns the implementation meter
func (s *Server) Meter() metric.MeterMust {
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
	for _, value := range s.closers {
		if value == closer {
			return
		}
	}
	switch closer.(type) {
	case Closer,
		ErrorCloser,
		CloserWithContext,
		ErrorCloserWithContext,
		Shutdowner,
		ErrorShutdowner,
		ShutdownerWithContext,
		ErrorShutdownerWithContext,
		Unsubscriber,
		ErrorUnsubscriber,
		UnsubscriberWithContext,
		ErrorUnsubscriberWithContext:
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

// AddHealthzer adds a probe to be called on healthz checks
func (s *Server) AddHealthzer(typ HealthzType, probe interface{}) {
	switch probe.(type) {
	case BoolHealthzer,
		BoolHealthzerWithContext,
		ErrorHealthzer,
		ErrorHealthzWithContext,
		ErrorPinger,
		ErrorPingerWithContext:
		s.probes[typ] = append(s.probes[typ], probe)
	default:
		s.l.Debug("not a healthz probe", log.FValue(fmt.Sprintf("%T", probe)))
	}
}

// AddHealthzers adds the given probes to be called on healthz checks
func (s *Server) AddHealthzers(typ HealthzType, probes ...interface{}) {
	for _, probe := range probes {
		s.AddHealthzer(typ, probe)
	}
}

// AddAlwaysHealthzers adds the probes to be called on any healthz checks
func (s *Server) AddAlwaysHealthzers(probes ...interface{}) {
	s.AddHealthzers(HealthzTypeAlways, probes...)
}

// AddStartupHealthzers adds the startup probes to be called on healthz checks
func (s *Server) AddStartupHealthzers(probes ...interface{}) {
	s.AddHealthzers(HealthzTypeStartup, probes...)
}

// AddLivenessHealthzers adds the liveness probes to be called on healthz checks
func (s *Server) AddLivenessHealthzers(probes ...interface{}) {
	s.AddHealthzers(HealthzTypeLiveness, probes...)
}

// AddReadinessHealthzers adds the readiness probes to be called on healthz checks
func (s *Server) AddReadinessHealthzers(probes ...interface{}) {
	s.AddHealthzers(HealthzTypeReadiness, probes...)
}

// IsCanceled returns true if the internal errgroup has been canceled
func (s *Server) IsCanceled() bool {
	return errors.Is(s.gCtx.Err(), context.Canceled)
}

// Healthz returns true if the server is running
func (s *Server) Healthz() error {
	if !s.running {
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

	defer s.ctxCancel()
	s.l.Info("starting keel server")

	// start services
	s.startService(s.services...)

	// set running
	defer func() {
		s.running = false
	}()
	s.running = true

	// wait for shutdown
	if err := s.g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.WithError(s.l, err).Error("service error")
	}

	s.l.Info("keel server stopped")
}

// startService starts the given services
func (s *Server) startService(services ...Service) {
	for _, service := range services {
		service := service
		s.g.Go(func() error {
			if err := service.Start(s.ctx); errors.Is(err, http.ErrServerClosed) {
				log.WithError(s.l, err).Debug("server has closed")
			} else if err != nil {
				log.WithError(s.l, err).Error("failed to start service")
				return err
			}
			return nil
		})
	}
}
