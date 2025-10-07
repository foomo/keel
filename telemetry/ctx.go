package telemetry

import (
	"context"
	"runtime"
	"strings"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Context struct {
	context.Context
}

func (c Context) Log() *zap.Logger {
	return Log(c.Context).WithOptions(zap.AddCallerSkip(1))
}

func (c Context) Trace(opts ...trace.SpanStartOption) (Context, trace.Span) {
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
	ctx, span := Tracer().Start(c.Context, name, append(opts, trace.WithAttributes(attrs...))...)
	return Ctx(ctx), span
}

func (c Context) Profile(handler func(ctx Context), labels ...string) {
	pyroscope.TagWrapper(c.Context, pyroscope.Labels(labels...), func(ctx context.Context) {
		handler(Ctx(ctx))
	})
}

func Ctx(ctx context.Context) Context {
	return Context{ctx}
}
