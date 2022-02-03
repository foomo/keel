package keel

import (
	"context"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
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

// WithContext option
func WithContext(ctx context.Context) Option {
	return func(inst *Server) {
		inst.ctx = ctx
	}
}

// WithShutdownSignals option
func WithShutdownSignals(shutdownSignals ...os.Signal) Option {
	return func(inst *Server) {
		inst.shutdownSignals = shutdownSignals
	}
}

// WithShutdownTimeout option
func WithShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(inst *Server) {
		inst.shutdownTimeout = shutdownTimeout
	}
}

// WithHTTPZapService option with default value
func WithHTTPZapService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.zap.enabled", enabled)() {
			inst.AddService(NewDefaultServiceHTTPZap())
		}
	}
}

// WithHTTPViperService option with default value
func WithHTTPViperService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.viper.enabled", enabled)() {
			inst.AddService(NewDefaultServiceHTTPViper())
		}
	}
}

// WithHTTPPrometheusService option with default value
func WithHTTPPrometheusService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.prometheus.enabled", enabled)() {
			inst.AddService(NewDefaultServiceHTTPPrometheus())
		}
	}
}
