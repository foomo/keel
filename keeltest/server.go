package keeltest

import (
	"context"
	"slices"
	"testing"

	testingx "github.com/foomo/go/testing"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Server is a lightweight test counterpart of the keel server that starts and
// closes [Service] instances within a test. Create one with [NewServer].
type Server struct {
	services      []Service
	serviceMap    map[string]Service
	ctx           context.Context
	meter         metric.Meter
	meterProvider metric.MeterProvider
	tracer        trace.Tracer
	traceProvider trace.TracerProvider
	l             *zap.Logger
	c             *viper.Viper
}

// NewServer returns a [Server] using the context of tb, a no-op logger, the
// global config and the global telemetry providers unless overridden by
// opts. [Server.Close] is registered as cleanup on tb.
func NewServer(tb testing.TB, opts ...Option) *Server {
	tb.Helper()

	inst := &Server{
		ctx:           tb.Context(),
		l:             zap.NewNop(),
		c:             config.Config(),
		meter:         telemetry.Meter(),
		tracer:        telemetry.Tracer(),
		meterProvider: telemetry.MeterProvider(),
		traceProvider: telemetry.TracerProvider(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	tb.Cleanup(inst.Close)

	return inst
}

// NewExampleServer returns a [Server] for use in Example functions. Its
// [Server.Close] must be called explicitly.
func NewExampleServer(opts ...Option) *Server {
	return NewServer(testingx.NewExampleTB(), opts...)
}

// Logger returns the server logger.
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Meter returns the server meter.
func (s *Server) Meter() metric.Meter {
	return s.meter
}

// Tracer returns the server tracer.
func (s *Server) Tracer() trace.Tracer {
	return s.tracer
}

// Config returns the server config.
func (s *Server) Config() *viper.Viper {
	return s.c
}

// Context returns the server context.
func (s *Server) Context() context.Context {
	return s.ctx
}

// AddServices adds services, see [Server.AddService].
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddService adds service unless it was already added. Services must be
// added before [Server.Start].
func (s *Server) AddService(service Service) {
	if slices.Contains(s.services, service) {
		return
	}

	s.services = append(s.services, service)
}

// GetService returns the service named name, or nil if there is none or
// [Server.Start] has not been called.
func (s *Server) GetService(name string) Service {
	if v, ok := s.serviceMap[name]; ok {
		return v
	}

	return nil
}

// Start starts all added services in order. Start errors are logged, not
// returned.
func (s *Server) Start() {
	s.serviceMap = make(map[string]Service, len(s.services))
	for _, service := range s.services {
		s.serviceMap[service.Name()] = service
		if err := service.Start(s.Context()); err != nil {
			log.WithError(s.l, err).Error("failed to start service")
		}
	}
}

// Close closes all registered services. It is registered as cleanup on the
// given testing.TB, but must be called explicitly in examples.
func (s *Server) Close() {
	for _, service := range s.services {
		if err := service.Close(context.WithoutCancel(s.Context())); err != nil {
			log.WithError(s.l, err).Error("failed to close service")
		}
	}
}
