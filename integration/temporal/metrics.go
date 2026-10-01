package keeltemporal

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
)

// NewMetricsHandler returns a Temporal metrics handler that records counters,
// gauges and timers (in seconds) with meter. Instrument creation errors are
// reported via [otel.Handle].
//
// Deprecated: Use [opentelemetry.NewMetricsHandler] instead.
func NewMetricsHandler(meter metric.Meter) client.MetricsHandler {
	return opentelemetry.NewMetricsHandler(opentelemetry.MetricsHandlerOptions{
		Meter:   meter,
		OnError: otel.Handle,
	})
}
