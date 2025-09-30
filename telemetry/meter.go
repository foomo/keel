package telemetry

import (
	"context"
	"encoding/json"
	"os"

	"github.com/foomo/keel/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	// DefaultHistogramBuckets units are selected for metrics in "seconds" unit
	DefaultHistogramBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25, 60, 120, 300, 600}
)

func Meter() metric.Meter {
	return otel.Meter(Name)
}

func NewNoopMeterProvider() (metric.MeterProvider, error) {
	provider := noop.NewMeterProvider()
	otel.SetMeterProvider(provider)
	return provider, nil
}

func NewStdOutMeterProvider(ctx context.Context, opts ...stdoutmetric.Option) (metric.MeterProvider, error) {
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		opts = append(opts, stdoutmetric.WithEncoder(enc))
	}
	if !env.GetBool("OTEL_EXPORTER_STDOUT_TIMESTAMP", true) {
		opts = append(opts, stdoutmetric.WithoutTimestamps())
	}

	exporter, err := stdoutmetric.New(opts...)
	if err != nil {
		return nil, err
	}

	resource := otelresource.NewSchemaless(
		semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
	)

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(resource),
	)

	otel.SetMeterProvider(provider)
	return provider, nil
}

func NewPrometheusMeterProvider() (metric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	resource := otelresource.NewSchemaless(
		semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
	)

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(resource),
	)

	otel.SetMeterProvider(provider)
	return provider, nil
}
