package telemetry

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns a [trace.Tracer] named [Name] from the global
// [TracerProvider]. The instrumentation scope version defaults to [Version];
// pass [trace.WithInstrumentationVersion] to override it. The schema URL
// defaults to the semconv version used by keel.
func Tracer(opts ...trace.TracerOption) trace.Tracer {
	return TracerProvider().Tracer(Name, append([]trace.TracerOption{trace.WithInstrumentationVersion(Version()), trace.WithSchemaURL(semconv.SchemaURL)}, opts...)...)
}
