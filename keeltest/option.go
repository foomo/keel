package keeltest

import (
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Option configures a [Server].
type Option func(inst *Server)

// WithLogger sets the server logger. Defaults to a no-op logger.
func WithLogger(l *zap.Logger) Option {
	return func(inst *Server) {
		inst.l = l
	}
}

// WithLogFields adds fields to the server logger. It must be applied after
// [WithLogger] to affect a custom logger.
func WithLogFields(fields ...zap.Field) Option {
	return func(inst *Server) {
		inst.l = inst.l.With(fields...)
	}
}

// WithConfig sets the server config. Defaults to [config.Config].
func WithConfig(c *viper.Viper) Option {
	return func(inst *Server) {
		inst.c = c
	}
}

// WithMeterProvider sets the meter provider. Defaults to the global meter
// provider. It does not change the meter returned by [Server.Meter].
func WithMeterProvider(v metric.MeterProvider) Option {
	return func(inst *Server) {
		inst.meterProvider = v
	}
}

// WithTracerProvider sets the tracer provider. Defaults to the global
// tracer provider. It does not change the tracer returned by
// [Server.Tracer].
func WithTracerProvider(v trace.TracerProvider) Option {
	return func(inst *Server) {
		inst.traceProvider = v
	}
}
