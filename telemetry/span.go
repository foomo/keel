package telemetry

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

// Deprecated: use StartFunc instead.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, spanName, opts...) //nolint:spancheck
}

func StartFunc(ctx context.Context) (context.Context, func(errs ...error)) {
	name := "FUNC"
	var attrs []attribute.KeyValue
	if pc, file, line, ok := runtime.Caller(1); ok {
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
	span.SetStatus(codes.Ok, "")
	return ctx, end(span)
}

func StartFuncRequest(r *http.Request) (*http.Request, func(errs ...error)) {
	ctx, end := StartFunc(r.Context())
	return r.WithContext(ctx), end
}

// Deprecated: use StartFunc instead.
func End(sp trace.Span, err error) {
	sp.SetStatus(codes.Ok, "")

	if err != nil {
		sp.RecordError(err)
		sp.SetStatus(codes.Error, err.Error())
	}

	sp.End()
}

func end(span trace.Span) func(errs ...error) {
	return func(errs ...error) {
		if len(errs) > 0 {
			err := errors.Join(errs...)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}
