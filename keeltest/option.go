package keeltest

import (
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Option func
type Option func(inst *Server)

// WithLogger option
func WithLogger(l *zap.Logger) Option {
	return func(inst *Server) {
		inst.l = l
	}
}

// WithLogFields option
func WithLogFields(fields ...zap.Field) Option {
	return func(inst *Server) {
		inst.l = inst.l.With(fields...)
	}
}

// WithConfig option
func WithConfig(c *viper.Viper) Option {
	return func(inst *Server) {
		inst.c = c
	}
}

func WithMeterProvider(v metric.MeterProvider) Option {
	return func(inst *Server) {
		inst.meterProvider = v
	}
}

func WithTracerProvider(v trace.TracerProvider) Option {
	return func(inst *Server) {
		inst.traceProvider = v
	}
}
