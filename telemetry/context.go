package telemetry

import (
	"context"
	"runtime/pprof"
	"time"

	goerrors "github.com/foomo/go/errors"
	foomosemconv "github.com/foomo/opentelemetry-go/semconv"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

// Context wraps a [context.Context] and provides span, log and profile
// helpers bound to it. It implements [context.Context]. Create one with
// [Ctx]; the zero value is not usable.
type Context struct {
	ctx context.Context
}

// Ctx returns a [Context] wrapping ctx.
func Ctx(ctx context.Context) Context {
	return Context{ctx}
}

// ------------------------------------------------------------------------------------------------
// ~ Log methods
// ------------------------------------------------------------------------------------------------

// LogDebug logs msg at debug level, see [Log].
func (c Context) LogDebug(msg string, kv ...attribute.KeyValue) {
	Log(c.ctx, zapcore.DebugLevel, msg, 1, kv...)
}

// LogInfo logs msg at info level, see [Log].
func (c Context) LogInfo(msg string, kv ...attribute.KeyValue) {
	Log(c.ctx, zapcore.InfoLevel, msg, 1, kv...)
}

// LogWarn logs msg at warn level, see [Log].
func (c Context) LogWarn(msg string, kv ...attribute.KeyValue) {
	Log(c.ctx, zapcore.WarnLevel, msg, 1, kv...)
}

// LogError logs msg at error level, see [Log].
func (c Context) LogError(msg string, kv ...attribute.KeyValue) {
	Log(c.ctx, zapcore.ErrorLevel, msg, 1, kv...)
}

// ------------------------------------------------------------------------------------------------
// ~ Context methods
// ------------------------------------------------------------------------------------------------

// Context returns the wrapped [context.Context].
func (c Context) Context() context.Context {
	return c.ctx
}

// WithCancel returns a copy of the context with a new Done channel.
// The returned context's Done channel is closed when the returned cancel function
// is called or when the parent context's Done channel is closed, whichever happens first.
func (c Context) WithCancel() (Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.ctx)
	return Ctx(ctx), cancel
}

// WithCancelCause returns a copy of the context with a new Done channel and
// a CancelCauseFunc instead of a CancelFunc. Calling cancel with a non-nil error
// (the "cause") records that error in the context; it can then be retrieved using Cause(ctx).
func (c Context) WithCancelCause() (Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(c.ctx)
	return Ctx(ctx), cancel
}

// WithDeadline returns a copy of the context with a deadline.
// The returned context's Done channel is closed when the deadline expires, when the
// returned cancel function is called, or when the parent context's Done channel is
// closed, whichever happens first.
func (c Context) WithDeadline(deadline time.Time) (Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.ctx, deadline)
	return Ctx(ctx), cancel
}

// WithTimeout returns a copy of the context with a timeout.
// It is equivalent to WithDeadline(time.Now().Add(timeout)).
func (c Context) WithTimeout(timeout time.Duration) (Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	return Ctx(ctx), cancel
}

// WithValue returns a copy of the context with the key-value pair associated.
func (c Context) WithValue(key, val any) Context {
	return Ctx(context.WithValue(c.ctx, key, val))
}

// WithoutCancel returns a copy of the context that is not canceled when
// parent is canceled. The returned context returns no Deadline or Err, and its
// Done channel is nil.
func (c Context) WithoutCancel() Context {
	return Ctx(context.WithoutCancel(c.ctx))
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key.
func (c Context) Value(key any) any {
	return c.ctx.Value(key)
}

// Deadline returns the time when work done on behalf of this context should be
// canceled. Deadline returns ok==false when no deadline is set.
func (c Context) Deadline() (deadline time.Time, ok bool) { //nolint:nonamedreturns
	return c.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context
// should be canceled. Done may return nil if this context can never be canceled.
func (c Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err returns a non-nil error value after Done is closed. Err returns Canceled
// if the context was canceled or DeadlineExceeded if the context's deadline passed.
func (c Context) Err() error {
	return c.ctx.Err()
}

// ------------------------------------------------------------------------------------------------
// ~ Trace methods
// ------------------------------------------------------------------------------------------------

// Span returns the current span of the context.
func (c Context) Span() trace.Span {
	return SpanFromContext(c.ctx)
}

// SetSpanDebug marks the current span as a debug span.
func (c Context) SetSpanDebug() {
	SetSpanDebug(c.Span())
}

// EndSpan ends the current span, see [EndSpan].
func (c Context) EndSpan(err error, opts ...trace.SpanEndOption) {
	EndSpan(c.Span(), err, opts...)
}

// DeferEndSpan ends the current span with the error err points to, so it
// can be used as defer ctx.DeferEndSpan(&err). See [DeferEndSpan].
func (c Context) DeferEndSpan(err *error, opts ...trace.SpanEndOption) {
	DeferEndSpan(c.Span(), err, opts...)
}

// SetSpanStatusOK sets the status of the current span to ok.
func (c Context) SetSpanStatusOK() {
	c.Span().SetStatus(codes.Ok, "")
}

// SetSpanStatusError sets the status of the current span to error with the
// given description.
func (c Context) SetSpanStatusError(description string) {
	c.Span().SetStatus(codes.Error, description)
	SetSpanStatusError(c.Span(), description)
}

// SetSpanName sets the name of the current span.
func (c Context) SetSpanName(name string) {
	SetSpanName(c.Span(), name)
}

// SetSpanAttributes sets attributes on the current span.
func (c Context) SetSpanAttributes(kv ...attribute.KeyValue) {
	SetSpanAttributes(c.Span(), kv...)
}

// RecordError records a non-nil err with kv and a stack trace on the
// current span and sets the span status to error with the root cause
// message. It does not log the error.
func (c Context) RecordError(err error, kv ...attribute.KeyValue) {
	if err != nil {
		sp := c.Span()
		sp.RecordError(err,
			trace.WithAttributes(kv...),
			trace.WithAttributes(CodeStacktrace(5, 1)),
		)
		sp.SetStatus(codes.Error, goerrors.Cause(err).Error())
	}
}

// RecordSpanError records a non-nil err with kv and a stack trace on the
// current span without changing the span status.
func (c Context) RecordSpanError(err error, kv ...attribute.KeyValue) {
	if err != nil {
		sp := c.Span()
		sp.RecordError(err,
			trace.WithAttributes(append(kv, CodeStacktrace(5, 1))...),
		)
	}
}

// AddSpanEvent adds an event with attributes kv to the current span.
func (c Context) AddSpanEvent(name string, kv ...attribute.KeyValue) {
	c.Span().AddEvent(name, trace.WithAttributes(kv...))
}

// AddSpanLink adds a link from the current span to parent.
func (c Context) AddSpanLink(parent trace.Span, attrs ...attribute.KeyValue) {
	sp := c.Span()
	sp.AddLink(trace.Link{
		SpanContext: parent.SpanContext(),
		Attributes:  attrs,
	})
}

// StartSpan starts a child span named after the calling package and returns
// a Context holding it, see [StartSpan].
func (c Context) StartSpan(opts ...trace.SpanStartOption) Context {
	ctx, _ := StartSpanWithSkip(c.ctx, 1, opts...)
	return Ctx(ctx)
}

// StartSpanWithNewRoot starts a span in a new trace, linked to the span of
// the context, and returns a Context holding it.
func (c Context) StartSpanWithNewRoot(opts ...trace.SpanStartOption) Context {
	ctx, _ := StartSpanWithSkip(c.ctx, 1, append(opts, trace.WithNewRoot(), trace.WithLinks(trace.LinkFromContext(c.ctx)))...)
	return Ctx(ctx)
}

// StartSpanWithProfile starts a span with attributes kv, runs handler within
// a profile named name (see [Context.StartProfile]) and ends the span.
func (c Context) StartSpanWithProfile(name string, handler func(ctx Context), kv ...attribute.KeyValue) {
	ctx, span := StartSpanWithSkip(c.ctx, 1, trace.WithAttributes(kv...))
	defer span.End()

	Ctx(ctx).StartProfile(name, handler, kv...)
}

// StartProfile runs handler with pyroscope labels for the profile name and
// kv applied and sets the profile name attribute on the current span.
func (c Context) StartProfile(name string, handler func(ctx Context), kv ...attribute.KeyValue) {
	attr := foomosemconv.ProfileName(name)
	c.Span().SetAttributes(attr)
	pyroscope.TagWrapper(c.ctx, PyroscopeLabels(append(kv, attr)...), func(ctx context.Context) {
		handler(Ctx(ctx))
	})
}

// SetProfileAttributes adds kv as pprof labels to the context and the
// current goroutine and returns the labeled Context.
func (c Context) SetProfileAttributes(kv ...attribute.KeyValue) Context {
	ctx := pprof.WithLabels(c.ctx, PyroscopeLabels(kv...))
	pprof.SetGoroutineLabels(ctx)

	return Ctx(ctx)
}
