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

// MeterProvider returns the global [metric.MeterProvider].
func MeterProvider() metric.MeterProvider {
	return otel.GetMeterProvider()
}

// NewNoopMeterProvider returns a no-op [metric.MeterProvider].
func NewNoopMeterProvider() metric.MeterProvider {
	return noop.NewMeterProvider()
}

// NewStdOutMeterProvider returns a meter provider that periodically exports
// metrics to stdout and installs it as the global meter provider. Output
// formatting is controlled by OTEL_EXPORTER_STDOUT_PRETTY_PRINT and
// OTEL_EXPORTER_STDOUT_TIMESTAMPS, both defaulting to true; opts are passed
// to the exporter.
func NewStdOutMeterProvider(ctx context.Context, opts ...stdoutmetric.Option) (metric.MeterProvider, error) {
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		opts = append(opts, stdoutmetric.WithEncoder(enc))
	}

	if !env.GetBool("OTEL_EXPORTER_STDOUT_TIMESTAMPS", true) {
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
// The provider is installed as the global meter provider.
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
// The provider is installed as the global meter provider.
func NewOTLPHTTPMeterProvider(ctx context.Context, opts ...otlpmetrichttp.Option) (metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newMeterProvider(ctx, sdkmetric.NewPeriodicReader(exporter))
}

// NewPrometheusMeterProvider returns a meter provider backed by the OTEL
// Prometheus exporter, which registers into prometheus.DefaultRegisterer, and
// installs it as the global meter provider.
func NewPrometheusMeterProvider(ctx context.Context) (metric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	return newMeterProvider(ctx, exporter)
}

// newMeterProvider creates an SDK meter provider with the default resource
// and reader r and installs it as the global meter provider. It starts host
// and runtime instrumentation if OTEL_METRICS_HOST_ENABLED or
// OTEL_METRICS_RUNTIME_ENABLED are set.
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
