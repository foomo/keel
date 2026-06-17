package telemetry

import (
	"context"
	"encoding/json"
	"os"

	"github.com/foomo/keel/env"
	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
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

// NewOTLPGRPCMeterProvider creates a push-based metric.MeterProvider that exports metrics
// over OTLP gRPC via a periodic reader. It configures the exporter from environment
// variables (e.g. endpoint, insecure transport) unless overridden by the given options.
// Push-based export suits short-lived workloads (e.g. Jobs) that exit before a scrape.
func NewOTLPGRPCMeterProvider(ctx context.Context, opts ...otlpmetricgrpc.Option) (metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newMeterProvider(ctx, sdkmetric.NewPeriodicReader(exporter))
}

// NewOTLPHTTPMeterProvider creates a push-based metric.MeterProvider that exports metrics
// over OTLP HTTP via a periodic reader. It configures the exporter from environment
// variables (e.g. endpoint, insecure transport) unless overridden by the given options.
// Push-based export suits short-lived workloads (e.g. Jobs) that exit before a scrape.
func NewOTLPHTTPMeterProvider(ctx context.Context, opts ...otlpmetrichttp.Option) (metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newMeterProvider(ctx, sdkmetric.NewPeriodicReader(exporter))
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
