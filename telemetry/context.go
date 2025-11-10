package telemetry

import (
	"context"
	"runtime/pprof"

	pkgsemconv "github.com/foomo/keel/semconv"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/internal/runtimeutil"
	"github.com/foomo/keel/log"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

type Context struct {
	context.Context
}

func Ctx(ctx context.Context) Context {
	return Context{ctx}
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// LogDebug logs a message at debug level.
func (c Context) LogDebug(msg string, kv ...attribute.KeyValue) {
	c.log(c.Context, zapcore.DebugLevel, msg, 1, kv...)
}

// LogInfo logs a message at info level.
func (c Context) LogInfo(msg string, kv ...attribute.KeyValue) {
	c.log(c.Context, zapcore.InfoLevel, msg, 1, kv...)
}

// LogWarn logs a message at warn level.
func (c Context) LogWarn(msg string, kv ...attribute.KeyValue) {
	c.log(c.Context, zapcore.WarnLevel, msg, 1, kv...)
}

// LogError logs a message at error level.
func (c Context) LogError(msg string, kv ...attribute.KeyValue) {
	c.log(c.Context, zapcore.ErrorLevel, msg, 1, kv...)
}

// Span returns the span from the context.
func (c Context) Span() trace.Span {
	return trace.SpanFromContext(c.Context)
}

func (c Context) SetSpanDebug() {
	c.Span().SetAttributes(attribute.Bool("debug.enabled", true))
}

// EndSpan ends the span.
func (c Context) EndSpan(err error, opts ...trace.SpanEndOption) {
	sp := c.Span()
	if sp.IsRecording() {
		if err != nil {
			sp.RecordError(err, trace.WithAttributes(semconv.CodeStacktrace(runtimeutil.StackTrace(3, 1))))
			sp.SetStatus(codes.Error, errors.Cause(err).Error())
		} else {
			c.SetSpanStatusOK()
		}

		sp.End(opts...)
	}
}

// DeferEndSpan is a helper so you can do `defer ctx.DeferEndSpan(&err)` instead of `defer func(){ ctx.EndSpan(err) }()`
func (c Context) DeferEndSpan(err *error, opts ...trace.SpanEndOption) {
	e := *err

	sp := c.Span()
	if sp.IsRecording() {
		if e != nil {
			sp.RecordError(e, trace.WithAttributes(pkgsemconv.CodeStacktrace(5, 2)))
			sp.SetStatus(codes.Error, errors.Cause(e).Error())
		} else {
			c.SetSpanStatusOK()
		}

		sp.End(opts...)
	}
}

// SetSpanStatusOK sets the status of the span to ok.
func (c Context) SetSpanStatusOK() {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetStatus(codes.Ok, "")
	}
}

// SetSpanStatusError sets the status of the span to error.
func (c Context) SetSpanStatusError(description string) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetStatus(codes.Error, description)
	}
}

// SetSpanName sets the name of the span.
func (c Context) SetSpanName(name string) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetName(name)
	}
}

// SetSpanAttributes sets the attributes of the span.
func (c Context) SetSpanAttributes(kv ...attribute.KeyValue) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetAttributes(kv...)
	}
}

// RecordError records an error on the span and logs it.
func (c Context) RecordError(err error, kv ...attribute.KeyValue) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.RecordError(err,
			trace.WithAttributes(kv...),
			trace.WithAttributes(pkgsemconv.CodeStacktrace(5, 1)),
		)
		sp.SetStatus(codes.Error, errors.Cause(err).Error())
	}
}

// RecordSpanError records an error on the span.
func (c Context) RecordSpanError(err error, kv ...attribute.KeyValue) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.RecordError(err,
			trace.WithAttributes(append(kv, pkgsemconv.CodeStacktrace(5, 1))...),
		)
	}
}

// AddSpanEvent adds an event to the span.
func (c Context) AddSpanEvent(name string, kv ...attribute.KeyValue) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.AddEvent(name, trace.WithAttributes(kv...))
	}
}

// StartSpan starts a span.
func (c Context) StartSpan(opts ...trace.SpanStartOption) Context {
	ctx, _ := c.startSpan("FUNC", 1, opts...)
	return ctx
}

// StartSpanWithProfile starts a span and profiles the handler.
func (c Context) StartSpanWithProfile(name string, handler func(ctx Context), kv ...attribute.KeyValue) {
	ctx, span := c.startSpan("FUNC", 1, trace.WithAttributes(kv...))
	defer span.End()

	ctx.StartProfile(name, handler, kv...)
}

// StartProfile starts a profile for the handler.
func (c Context) StartProfile(name string, handler func(ctx Context), kv ...attribute.KeyValue) {
	pyroscope.TagWrapper(c.Context, PyroscopeLabels(append(kv, pkgsemconv.ProfileName(name))...), func(ctx context.Context) {
		handler(Ctx(ctx))
	})
}

// SetProfileAttributes sets the labels for the profile.
func (c Context) SetProfileAttributes(kv ...attribute.KeyValue) Context {
	ctx := pprof.WithLabels(c.Context, PyroscopeLabels(kv...))
	pprof.SetGoroutineLabels(ctx)

	return Ctx(ctx)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c Context) startSpan(prefix string, skip int, opts ...trace.SpanStartOption) (Context, trace.Span) {
	name := prefix

	if shortName, _, file, line, ok := runtimeutil.Caller(skip + 1); ok {
		name += " " + shortName
		opts = append(opts, trace.WithAttributes(
			semconv.CodeFunctionName(shortName),
			semconv.CodeLineNumber(line),
			semconv.CodeFilePath(file),
		))
	}

	ctx, span := Tracer().Start(c.Context, name, opts...) //nolint:spancheck

	return Ctx(ctx), span //nolint:spancheck
}

func (c Context) log(ctx context.Context, lvl zapcore.Level, msg string, skip int, kv ...attribute.KeyValue) {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		kv = append(kv,
			pkgsemconv.TraceID(spanCtx.TraceID().String()),
			pkgsemconv.SpanID(spanCtx.SpanID().String()),
		)
	}

	kv = append(kv, pkgsemconv.CodeCaller(skip+1)...)

	zap.L().WithOptions(zap.WithCaller(false)).Log(lvl, msg, log.Attributes(kv...)...)
}
