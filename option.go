package keel

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Option func
type Option func(inst *Server)

func WithLogger(l *zap.Logger) Option {
	return func(inst *Server) {
		inst.l = l
	}
}

func WithLogFields(fields ...zap.Field) Option {
	return func(inst *Server) {
		inst.l = inst.l.With(fields...)
	}
}

func WithConfig(c *viper.Viper) Option {
	return func(inst *Server) {
		inst.c = c
	}
}

func WithContext(ctx context.Context) Option {
	return func(inst *Server) {
		inst.ctx = ctx
	}
}

func WithShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(inst *Server) {
		inst.shutdownTimeout = shutdownTimeout
	}
}
