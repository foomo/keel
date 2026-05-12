package telemetry

import (
	"context"

	"github.com/foomo/keel/log"
	foomosemconv "github.com/foomo/opentelemetry-go/semconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func LogWarn(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.WarnLevel, msg, 1, kv...)
}

func LogError(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.ErrorLevel, msg, 1, kv...)
}

func LogDebug(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.DebugLevel, msg, 1, kv...)
}

func LogInfo(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.InfoLevel, msg, 1, kv...)
}

func Log(ctx context.Context, lvl zapcore.Level, msg string, skip int, kv ...attribute.KeyValue) {
	if !zap.L().Core().Enabled(lvl) {
		return
	}

	attrs := make([]attribute.KeyValue, 0, len(kv)+5)
	attrs = append(attrs, kv...)

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		attrs = append(attrs,
			foomosemconv.TraceID(spanCtx.TraceID().String()),
			foomosemconv.SpanID(spanCtx.SpanID().String()),
		)
	}

	attrs = append(attrs, CodeCaller(skip+1)...)

	zap.L().WithOptions(zap.WithCaller(false)).Log(lvl, msg, log.Attributes(attrs...)...)
}
