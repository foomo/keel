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

	return inst
}

func NewExampleServer(opts ...Option) *Server {
	return NewServer(testingx.NewExampleTB(), opts...)
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

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	if slices.Contains(s.services, service) {
		return
	}

	s.services = append(s.services, service)
}

func (s *Server) GetService(name string) Service {
	if v, ok := s.serviceMap[name]; ok {
		return v
	}

	return nil
}

// Start starts all registered services
func (s *Server) Start() {
	s.serviceMap = make(map[string]Service, len(s.services))
	for _, service := range s.services {
		s.serviceMap[service.Name()] = service
		if err := service.Start(s.Context()); err != nil {
			s.l.Error("failed to start service", log.FError(err))
		}
	}
}
