package telemetrytest

import (
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// NewTestTraceProvider returns a tracer provider that records all spans in
// the returned span recorder and installs it as the global tracer provider.
func NewTestTraceProvider() (*tracetest.SpanRecorder, trace.TracerProvider) {
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	otel.SetTracerProvider(tracerProvider)

	return spanRecorder, tracerProvider
}
