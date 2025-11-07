package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

// Meter returns a metric.Meter instance with the provided options, using the default MeterProvider and the defined Name.
func Meter(opts ...metric.MeterOption) metric.Meter {
	return MeterProvider().Meter(Name, opts...)
}
