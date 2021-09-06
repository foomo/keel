package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	otelglobal "go.opentelemetry.io/otel/metric/global"
)

func Meter(instrumentationName string, opts ...metric.MeterOption) metric.MeterMust {
	return metric.Must(otelglobal.Meter(instrumentationName, opts...))
}
