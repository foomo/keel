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
	meter           metric.MeterMust
	meterProvider   metric.MeterProvider
	tracer          trace.Tracer
	traceProvider   trace.TracerProvider
	shutdownSignals []os.Signal
	shutdownTimeout time.Duration
	closers         []interface{}
	probeHandlers   []ProbeHandlers
	ctx             context.Context
	l               *zap.Logger
	c               *viper.Viper
}

func NewServer(opts ...Option) *Server {
	inst := &Server{
		shutdownTimeout: 5 * time.Second,
		shutdownSignals: []os.Signal{os.Interrupt, syscall.SIGTERM},
		ctx:             context.Background(),
		c:               config.Config(),
		l:               log.Logger(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	var err error

	// set otel error handler
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
	for index, value := range s.services {
		if value == service {
			return
		} else if value.Name() == service.Name() {
			s.services[index] = service
			return
		}
	}
	s.services = append(s.services, service)
}

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddCloser adds a closer to be called on shutdown
func (s *Server) AddCloser(closer interface{}) {
	switch closer.(type) {
	case Closer,
		CloserFn,
		ErrorCloser,
		ErrorCloserFn,
		CloserWithContext,
		CloserWithContextFn,
		ErrorCloserWithContext,
		ErrorCloserWithContextFn,
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
		s.l.Warn("unable to add closer")
	}
}

// AddClosers adds a closer to be called on shutdown
func (s *Server) AddClosers(closers ...interface{}) {
	for _, closer := range closers {
		s.AddCloser(closer)
	}
}

// AddCloser adds a closer to be called on shutdown
func (s *Server) AddProbeHandlers(health interface{}, probe ProbeType) {
	switch health.(type) {
	case Health,
		HealthFn,
		ErrorHealth,
		ErrorHealthFn,
		HealthWithContext,
		HealthWithContextFn,
		ErrorHealthWithContext:
		s.probeHandlers = append(s.probeHandlers, ProbeHandlers{
			probeType: string(probe),
			handler:   health,
		})
	default:
		s.l.Warn("unable to add probe handlers")
	}
}

// Run runs the server
func (s *Server) Run() {
	s.l.Info("starting server")

	ctx, stop := signal.NotifyContext(s.ctx, s.shutdownSignals...)
	defer stop()

	g, gctx := errgroup.WithContext(ctx)

	if len(s.probeHandlers) > 0 {
		handler := CreateProbeHandlers(s)
		s.AddService(NewServiceHTTP(log.Logger(), DefaultServiceHTTPProbesName, DefaultServiceHTTPProbesAddr, handler))
	}

	for _, service := range s.services {
		service := service
		g.Go(func() error {
			if err := service.Start(s.ctx); errors.Is(err, http.ErrServerClosed) {
				log.WithError(s.l, err).Debug("server has closed")
			} else if err != nil {
				log.WithError(s.l, err).Error("failed to start service")
				return err
			}
			return nil
		})
		// register started service
		s.AddCloser(service)
	}

	// gracefully shutdown servers
	g.Go(func() error {
		<-gctx.Done()
		s.l.Debug("gracefully stopping closers...")

		timeoutCtx, timeoutCancel := context.WithTimeout(
			context.Background(),
			s.shutdownTimeout,
		)
		defer timeoutCancel()

		// append internal closers
		closers := append(s.closers, s.traceProvider, s.meterProvider) //nolint:gocritic

		for _, closer := range closers {
			switch c := closer.(type) {
			case CloserFn:
				c()
			case Closer:
				c.Close()
			case ErrorCloserFn:
				if err := c(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorCloser")
					continue
				}
			case ErrorCloser:
				if err := c.Close(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorCloser")
					continue
				}
			case CloserWithContextFn:
				c(timeoutCtx)
			case CloserWithContext:
				c.Close(timeoutCtx)
			case ErrorCloserWithContextFn:
				if err := c(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorCloserWithContext")
					continue
				}
			case ErrorCloserWithContext:
				if err := c.Close(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorCloserWithContext")
					continue
				}
			case Shutdowner:
				c.Shutdown()
			case ErrorShutdowner:
				if err := c.Shutdown(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorShutdowner")
					continue
				}
			case ShutdownerWithContext:
				c.Shutdown(timeoutCtx)
			case ErrorShutdownerWithContext:
				if err := c.Shutdown(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorShutdownerWithContext")
					continue
				}
			case Unsubscriber:
				c.Unsubscribe()
			case ErrorUnsubscriber:
				if err := c.Unsubscribe(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorUnsubscriber")
					continue
				}
			case UnsubscriberWithContext:
				c.Unsubscribe(timeoutCtx)
			case ErrorUnsubscriberWithContext:
				if err := c.Unsubscribe(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorUnsubscriberWithContext")
					continue
				}
			}
			s.l.Info("stopped registered closer", log.FName(fmt.Sprintf("%T", closer)))
		}
		return gctx.Err()
	})

	// wait for shutdown
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.WithError(s.l, err).Error("service error")
	}

	s.l.Info("graceful shutdown complete")
}
