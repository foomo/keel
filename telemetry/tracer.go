package telemetry

import (
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns a [trace.Tracer] named [Name] from the global
// [TracerProvider]. The instrumentation scope version defaults to [Version];
// pass [trace.WithInstrumentationVersion] to override it.
func Tracer(opts ...trace.TracerOption) trace.Tracer {
	return TracerProvider().Tracer(Name, append([]trace.TracerOption{trace.WithInstrumentationVersion(Version())}, opts...)...)
}
