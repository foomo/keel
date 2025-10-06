package telemetry

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"strings"

	errors2 "github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Deprecated: use StartCode instead.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, spanName, opts...) //nolint:spancheck
}

// Deprecated: use StartCode instead.
func End(sp trace.Span, err error) {
	if err != nil {
		sp.RecordError(err)
		sp.SetStatus(codes.Error, err.Error())
	} else {
		sp.SetStatus(codes.Ok, "")
	}
	sp.End()
}

func StartCode(ctx context.Context) (context.Context, func(errs ...error)) {
	return startCode(ctx)
}

func StartCodeRequest(r *http.Request) (*http.Request, func(errs ...error)) {
	ctx, end := startCode(r.Context())
	return r.WithContext(ctx), end
}

func startCode(ctx context.Context) (context.Context, func(errs ...error)) {
	name := "CODE"
	var attrs []attribute.KeyValue
	if pc, file, line, ok := runtime.Caller(2); ok {
		attrs = append(attrs,
			semconv.CodeLineNumber(line),
			semconv.CodeFilePath(file),
		)
		if details := runtime.FuncForPC(pc); details != nil {
			funcName := details.Name()
			attrs = append(attrs, semconv.CodeFunctionName(funcName))
			lastSlash := strings.LastIndexByte(funcName, '/')
			if lastSlash < 0 {
				lastSlash = 0
			}
			lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash
			name += " " + funcName[lastDot+1:]
		}
	}

	ctx, span := Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
	return ctx, end(span)
}

func end(span trace.Span) func(errs ...error) {
	return func(errs ...error) {
		if span.IsRecording() {
			var err error
			if len(errs) > 1 {
				err = errors.Join(errs...)
			} else if len(errs) == 1 {
				err = errs[0]
			}
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				logFromSpanContext(span.SpanContext()).WithOptions(zap.AddCallerSkip(1)).With(zap.Error(err)).Error(errors2.Cause(err).Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}
			span.End()
		}
	}
}
