package keeltest

import (
	"context"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
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

func NewServer(opts ...Option) *Server {
	inst := &Server{
		ctx: context.Background(),
		c:   config.Config(),
		l:   zap.L(),
	}

	{
		inst.meterProvider = noop.NewMeterProvider()
		inst.meter = inst.meterProvider.Meter("github.com/foomo/keel")
		traceProfiver, err := telemetry.NewNoopTraceProvider()
		log.Must(inst.l, err, "failed to create noop trace provider")
		inst.traceProvider = traceProfiver
		inst.tracer = inst.traceProvider.Tracer("github.com/foomo/keel")
	}

	for _, opt := range opts {
		opt(inst)
	}

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

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	for _, value := range s.services {
		if value == service {
			return
		}
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
