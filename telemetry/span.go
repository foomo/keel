package telemetry

import (
	"context"
	"path"

	goruntime "github.com/foomo/go/runtime"
	keelsemconv "github.com/foomo/keel/semconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
)

// Deprecated: use StartSpan instead.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartSpanWithSkip(ctx, 1, opts...)
}

// StartSpan starts a new span with the given name and options.
func StartSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartSpanWithSkip(ctx, 1, opts...)
}

// StartDebugSpan starts a new span with the given name and options and adds the debug attr.
func StartDebugSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartSpanWithSkip(ctx, 1, append(opts, trace.WithAttributes(keelsemconv.DebugEnabled(true)))...)
}

// SpanFromContext returns the span from the context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the span.
func AddSpanEvent(sp trace.Span, name string, opts ...trace.EventOption) {
	sp.AddEvent(name, opts...)
}

// AddSpanLink adds a link to the span.
func AddSpanLink(sp, parent trace.Span, attrs ...attribute.KeyValue) {
	sp.AddLink(trace.Link{
		SpanContext: parent.SpanContext(),
		Attributes:  attrs,
	})
}

// SetSpanAttributes sets attributes on the span.
func SetSpanAttributes(sp trace.Span, attrs ...attribute.KeyValue) {
	sp.SetAttributes(attrs...)
}

// SetSpanName sets the name of the span.
func SetSpanName(sp trace.Span, name string) {
	sp.SetName(name)
}

// IsSpanRecording returns true if the span is recording.
func IsSpanRecording(sp trace.Span) bool {
	return sp.IsRecording()
}

// SetSpanDebug sets the span debug attribute.
func SetSpanDebug(sp trace.Span) {
	sp.SetAttributes(keelsemconv.DebugEnabled(true))
}

// SetSpanStatusOK sets the status of the span to ok.
func SetSpanStatusOK(sp trace.Span) {
	sp.SetStatus(codes.Ok, "")
}

// SetSpanStatusError sets the status of the span to error.
func SetSpanStatusError(sp trace.Span, description string) {
	sp.SetStatus(codes.Error, description)
}

// Deprecated: use EndSpan instead.
func End(sp trace.Span, err error) {
	if err != nil {
		sp.RecordError(err, trace.WithAttributes(CodeStacktrace(3, 0)))
		sp.SetStatus(codes.Error, rootCause(err).Error())
	}

	sp.End()
}

// EndSpan ends the span.
func EndSpan(sp trace.Span, err error, opts ...trace.SpanEndOption) {
	if err != nil {
		sp.RecordError(err, trace.WithAttributes(CodeStacktrace(3, 0)))
		sp.SetStatus(codes.Error, rootCause(err).Error())
	}

	sp.End(opts...)
}

// DeferEndSpan is a helper, so you can do `defer ctx.DeferEndSpan(&err)` instead of `defer func(){ ctx.EndSpan(err) }()`
func DeferEndSpan(sp trace.Span, err *error, opts ...trace.SpanEndOption) {
	if err == nil {
		EndSpan(sp, nil, opts...)

		return
	}

	EndSpan(sp, *err, opts...)
}

// StartSpanWithSkip starts a span with skip.
// It's only supposed to be used internally or by other telemetry packages
func StartSpanWithSkip(ctx context.Context, skip int, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	name := "runtime.go"

	if fr := goruntime.CallFrame(skip + 1); !fr.Zero() {
		name = path.Base(fr.Pkg)
		opts = append(opts, trace.WithAttributes(
			semconv.CodeFunctionName(fr.Name()),
			semconv.CodeLineNumber(fr.Line),
			semconv.CodeFilePath(fr.File),
		))
	}

	return Tracer().Start(ctx, name, opts...) //nolint:spancheck
}
