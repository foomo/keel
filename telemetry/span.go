package telemetry

import (
	"context"
	"path"

	goerrors "github.com/foomo/go/errors"
	goruntime "github.com/foomo/go/runtime"
	keelsemconv "github.com/foomo/keel/semconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// Start starts a span like [StartSpan]; spanName is ignored.
//
// Deprecated: Use [StartSpan] instead.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartSpanWithSkip(ctx, 1, opts...)
}

// StartSpan starts a span named <package>.<function> after the caller (e.g.
// "keelmongo.(*Collection).Find"), annotated with the caller's function, file
// and line, from the global [Tracer].
func StartSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartSpanWithSkip(ctx, 1, opts...)
}

// StartDebugSpan is like [StartSpan] but marks the span as a debug span.
func StartDebugSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartSpanWithSkip(ctx, 1, append(opts, trace.WithAttributes(keelsemconv.DebugEnabled(true)))...)
}

// SpanFromContext returns the current span of ctx, or a non-recording span
// if there is none.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to sp.
func AddSpanEvent(sp trace.Span, name string, opts ...trace.EventOption) {
	sp.AddEvent(name, opts...)
}

// AddSpanLink adds a link from sp to parent with attrs.
func AddSpanLink(sp, parent trace.Span, attrs ...attribute.KeyValue) {
	sp.AddLink(trace.Link{
		SpanContext: parent.SpanContext(),
		Attributes:  attrs,
	})
}

// SetSpanAttributes sets attrs on sp.
func SetSpanAttributes(sp trace.Span, attrs ...attribute.KeyValue) {
	sp.SetAttributes(attrs...)
}

// SetSpanName sets the name of sp.
func SetSpanName(sp trace.Span, name string) {
	sp.SetName(name)
}

// IsSpanRecording reports whether sp is recording.
func IsSpanRecording(sp trace.Span) bool {
	return sp.IsRecording()
}

// SetSpanDebug marks sp as a debug span.
func SetSpanDebug(sp trace.Span) {
	sp.SetAttributes(keelsemconv.DebugEnabled(true))
}

// SetSpanStatusOK sets the status of sp to ok.
//
// Leave the status unset on success; backends treat unset as ok. Statuses
// override in the order Ok > Error > Unset: once set to ok, later error
// statuses are ignored. Use it only to explicitly mark an operation as
// successful despite recorded errors, e.g. a retry that succeeded after
// failed attempts set the status to error.
func SetSpanStatusOK(sp trace.Span) {
	sp.SetStatus(codes.Ok, "")
}

// SetSpanStatusError sets the status of sp to error with description.
func SetSpanStatusError(sp trace.Span, description string) {
	sp.SetStatus(codes.Error, description)
}

// End ends sp like [EndSpan].
//
// Deprecated: Use [EndSpan] instead.
func End(sp trace.Span, err error) {
	if err != nil {
		sp.RecordError(err, trace.WithAttributes(CodeStacktrace(3, 0)))
		sp.SetStatus(codes.Error, goerrors.Cause(err).Error())
		sp.SetAttributes(semconv.ErrorType(err))
	}

	sp.End()
}

// EndSpan ends sp. A non-nil err is recorded with a stack trace and sets the
// span status to error with the root cause message and error.type.
func EndSpan(sp trace.Span, err error, opts ...trace.SpanEndOption) {
	if err != nil {
		sp.RecordError(err, trace.WithAttributes(CodeStacktrace(3, 0)))
		sp.SetStatus(codes.Error, goerrors.Cause(err).Error())
		sp.SetAttributes(semconv.ErrorType(err))
	}

	sp.End(opts...)
}

// DeferEndSpan ends sp with the error err points to, so it can be used as
// defer DeferEndSpan(sp, &err). A nil err pointer ends the span without
// error.
func DeferEndSpan(sp trace.Span, err *error, opts ...trace.SpanEndOption) {
	if err == nil {
		EndSpan(sp, nil, opts...)

		return
	}

	EndSpan(sp, *err, opts...)
}

// StartSpanWithSkip is like [StartSpan] but uses the caller skip frames above
// the caller of StartSpanWithSkip for the span name and code attributes. It
// is intended for telemetry helper packages.
func StartSpanWithSkip(ctx context.Context, skip int, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	name := "runtime.go"

	if fr := goruntime.CallFrame(skip + 1); !fr.Zero() {
		name = path.Base(fr.Pkg) + "." + fr.Short()
		opts = append(opts, trace.WithAttributes(
			semconv.CodeFunctionName(fr.Name()),
			semconv.CodeLineNumber(fr.Line),
			semconv.CodeFilePath(fr.File),
		))
	}

	return Tracer().Start(ctx, name, opts...) //nolint:spancheck
}
