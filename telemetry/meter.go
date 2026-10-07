package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Meter returns a [metric.Meter] named [Name] from the global
// [MeterProvider]. The instrumentation scope version defaults to [Version];
// pass [metric.WithInstrumentationVersion] to override it. The schema URL
// defaults to the semconv version used by keel.
func Meter(opts ...metric.MeterOption) metric.Meter {
	return MeterProvider().Meter(Name, append([]metric.MeterOption{metric.WithInstrumentationVersion(Version()), metric.WithSchemaURL(semconv.SchemaURL)}, opts...)...)
}
