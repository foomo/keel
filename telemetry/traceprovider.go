package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/foomo/keel/env"
)

// TracerProvider returns the global [trace.TracerProvider].
func TracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

// NewNoopTraceProvider creates and returns a no-op tracer provider.
func NewNoopTraceProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

// NewStdOutTraceProvider creates a new stdout trace provider with optional configurations from the environment variables.
// It uses pretty print and timestamps by default unless overridden by the respective environment variables.
// The provider is installed as the global tracer provider.
func NewStdOutTraceProvider(ctx context.Context) (trace.TracerProvider, error) {
	var opts []stdouttrace.Option
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	if !env.GetBool("OTEL_EXPORTER_STDOUT_TIMESTAMPS", true) {
		opts = append(opts, stdouttrace.WithoutTimestamps())
	}

	exporter, err := stdouttrace.New(opts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(ctx, exporter)
}

// NewOTLPHTTPTraceProvider creates an OTLP HTTP trace provider using the given context and options.
// It configures the provider based on environment variables, such as endpoint and insecure transport.
// The provider is installed as the global tracer provider.
func NewOTLPHTTPTraceProvider(ctx context.Context, opts ...otlptracehttp.Option) (trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(ctx, exporter)
}

// NewOTLPGRPCTraceProvider creates a new trace provider configured for OTLP over gRPC with optional settings.
// It uses environment variables for settings like endpoint and insecure mode if not provided explicitly.
// The provider is installed as the global tracer provider.
func NewOTLPGRPCTraceProvider(ctx context.Context, opts ...otlptracegrpc.Option) (trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(ctx, exporter)
}

// newTracerProvider creates an SDK tracer provider batching to exp with the
// default resource and a parent based sampler using OTEL_TRACE_RATIO
// (default 1), and installs it as the global tracer provider.
func newTracerProvider(ctx context.Context, exp sdktrace.SpanExporter) (trace.TracerProvider, error) {
	resource, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(env.GetFloat64("OTEL_TRACE_RATIO", 1)),
			),
		),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}
