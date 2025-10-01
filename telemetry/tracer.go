package telemetry

import (
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns a trace.Tracer instance configured with optional TracerOptions.
func Tracer(opts ...trace.TracerOption) trace.Tracer {
	return TracerProvider().Tracer(Name, opts...)
}
