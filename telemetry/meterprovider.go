package telemetry

import (
	"context"
	"encoding/json"
	"os"

	"github.com/foomo/keel/env"
	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// MeterProvider returns the default metric.MeterProvider instance for creating meters.
func MeterProvider() metric.MeterProvider {
	return otel.GetMeterProvider()
}

// NewNoopMeterProvider returns a no-op metric.MeterProvider and sets it as the global MeterProvider.
func NewNoopMeterProvider() metric.MeterProvider {
	return noop.NewMeterProvider()
}

// NewStdOutMeterProvider creates a new MeterProvider that exports metrics to standard output with configurable options.
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

	reader := sdkmetric.NewPeriodicReader(exporter)

	return newMeterProvider(ctx, reader)
}

// NewPrometheusMeterProvider initializes and returns a Prometheus-based metric.MeterProvider with default configuration.
func NewPrometheusMeterProvider(ctx context.Context) (metric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	return newMeterProvider(ctx, exporter)
}

func newMeterProvider(ctx context.Context, r sdkmetric.Reader) (metric.MeterProvider, error) {
	if env.GetBool("OTEL_METRICS_HOST_ENABLED", false) {
		if err := otelhost.Start(); err != nil {
			return nil, err
		}
	}

	if env.GetBool("OTEL_METRICS_RUNTIME_ENABLED", false) {
		if err := otelruntime.Start(); err != nil {
			return nil, err
		}
	}

	resource, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(r),
		sdkmetric.WithResource(resource),
	)

	otel.SetMeterProvider(provider)

	return provider, nil
}
