package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Log(ctx context.Context, opts ...zap.Option) *zap.Logger {
	return logFromSpanContext(trace.SpanContextFromContext(ctx), opts...)
}

func logFromSpanContext(ctx trace.SpanContext, opts ...zap.Option) *zap.Logger {
	var fields []zapcore.Field
	if ctx.IsValid() {
		fields = append(fields,
			zap.String("trace_id", ctx.TraceID().String()),
			zap.String("span_id", ctx.SpanID().String()),
		)
	}

	return zap.L().With(fields...).WithOptions(opts...)
}
