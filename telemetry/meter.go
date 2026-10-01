package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

// Meter returns a [metric.Meter] named [Name] from the global
// [MeterProvider]. The instrumentation scope version defaults to [Version];
// pass [metric.WithInstrumentationVersion] to override it.
func Meter(opts ...metric.MeterOption) metric.Meter {
	return MeterProvider().Meter(Name, append([]metric.MeterOption{metric.WithInstrumentationVersion(Version())}, opts...)...)
}
