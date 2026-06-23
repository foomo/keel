package keel

import (
	"context"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Runtime is the shared accessor surface of Server and Job. Integrations should
// accept a Runtime instead of a concrete *Server so the same instrumented helper
// works under a long-lived Server or a short-lived Job.
type Runtime interface {
	// Logger returns the structured logger.
	Logger() *zap.Logger
	// Config returns the configuration.
	Config() *viper.Viper
	// Context returns the root context.
	Context() context.Context
	// Meter returns the OpenTelemetry meter.
	Meter() metric.Meter
	// Tracer returns the OpenTelemetry tracer.
	Tracer() trace.Tracer
	// AddCloser registers a closer to be called on shutdown/finalization.
	AddCloser(closer any)
	// AddClosers registers the given closers to be called on shutdown/finalization.
	AddClosers(closers ...any)
}

var (
	_ Runtime = (*Server)(nil)
	_ Runtime = (*Job)(nil)
)
