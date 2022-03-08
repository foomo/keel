package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/foomo/keel/env"
)

func Tracer() trace.Tracer {
	return TraceProvider().Tracer(TracerName)
}

func TraceProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

func NewNoopTraceProvider() (trace.TracerProvider, error) {
	tracerProvider := trace.NewNoopTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	return tracerProvider, nil
}

func NewStdOutTraceProvider(ctx context.Context) (trace.TracerProvider, error) {
	var exportOpts []stdouttrace.Option
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		exportOpts = append(exportOpts, stdouttrace.WithPrettyPrint())
	}

	exporter, err := stdouttrace.New(exportOpts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(exporter)
}

func NewOTLPHTTPTraceProvider(ctx context.Context, opts ...otlptracehttp.Option) (trace.TracerProvider, error) {
	if env.GetBool("OTEL_EXPORTER_OTLP_INSECURE", false) {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	opts = append(opts,
		otlptracehttp.WithEndpoint(env.MustGet("OTEL_EXPORTER_OTLP_ENDPOINT")),
	)

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(exporter)
}

func NewOTLPGRPCTraceProvider(ctx context.Context, endpoint string, opts ...otlptracegrpc.Option) (trace.TracerProvider, error) {
	if env.GetBool("OTEL_EXPORTER_OTLP_INSECURE", false) {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	opts = append(opts,
		otlptracegrpc.WithEndpoint(env.MustGet("OTEL_EXPORTER_OTLP_ENDPOINT")),
	)

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(exporter)
}

func newTracerProvider(e sdktrace.SpanExporter) (trace.TracerProvider, error) {
	resource := otelresource.NewSchemaless(
		semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(e),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracerProvider)
	return tracerProvider, nil
}
