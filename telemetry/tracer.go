package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/foomo/keel/env"
)

var traceMiddlewares []func(t trace.TracerProvider) trace.TracerProvider

func setTracerProvider(t trace.TracerProvider) {
	for _, middleware := range traceMiddlewares {
		t = middleware(t)
	}
	otel.SetTracerProvider(t)
}

func AddTraceMiddleware(m func(t trace.TracerProvider) trace.TracerProvider) {
	traceMiddlewares = append(traceMiddlewares, m)
}

func Tracer() trace.Tracer {
	return TraceProvider().Tracer(Name)
}

func TraceProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

func NewNoopTraceProvider() (trace.TracerProvider, error) {
	tracerProvider := noop.NewTracerProvider()
	setTracerProvider(tracerProvider)
	return tracerProvider, nil
}

func NewStdOutTraceProvider(ctx context.Context) (trace.TracerProvider, error) {
	var exportOpts []stdouttrace.Option
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		exportOpts = append(exportOpts, stdouttrace.WithPrettyPrint())
	}
	if !env.GetBool("OTEL_EXPORTER_STDOUT_TIMESTAMPS", true) {
		exportOpts = append(exportOpts, stdouttrace.WithoutTimestamps())
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
	if value := env.Get("OTEL_EXPORTER_OTLP_ENDPOINT", ""); value != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(value))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return newTracerProvider(exporter)
}

func NewOTLPGRPCTraceProvider(ctx context.Context, opts ...otlptracegrpc.Option) (trace.TracerProvider, error) {
	if env.GetBool("OTEL_EXPORTER_OTLP_INSECURE", false) {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if value := env.Get("OTEL_EXPORTER_OTLP_ENDPOINT", ""); value != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(value))
	}

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
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(env.GetFloat64("OTEL_TRACE_RATIO", 1)),
			),
		),
	)

	setTracerProvider(tracerProvider)
	return tracerProvider, nil
}
