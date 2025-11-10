package telemetry

import (
	"context"

	"github.com/foomo/keel/internal/runtimeutil"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func logger(ctx context.Context, skip int, opts ...zap.Option) *zap.Logger {
	return zap.L().
		WithOptions(append(opts, zap.WithCaller(false))...).
		With(logSpanFields(trace.SpanContextFromContext(ctx))...)
}

func logSpanFields(ctx trace.SpanContext, opts ...zap.Option) []zapcore.Field {
	if ctx.IsValid() {
		return []zapcore.Field{
			zap.String("trace_id", ctx.TraceID().String()),
			zap.String("span_id", ctx.SpanID().String()),
		}
	}

	return nil
}

func logCallerFields(skip int) []zapcore.Field {
	if shortName, _, file, line, ok := runtimeutil.Caller(skip + 1); ok {
		return []zapcore.Field{
			zap.String("code_function_name", shortName),
			zap.String("code_file_path", file),
			zap.Int("span_lint_number", line),
		}
	}

	return nil
}
