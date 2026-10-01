package telemetry

import (
	"context"
	"path"

	goruntime "github.com/foomo/go/runtime"
	"github.com/foomo/keel/log"
	foomosemconv "github.com/foomo/opentelemetry-go/semconv"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogWarn logs msg at warn level, see [Log].
func LogWarn(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.WarnLevel, msg, 1, kv...)
}

// LogError logs msg at error level, see [Log].
func LogError(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.ErrorLevel, msg, 1, kv...)
}

// LogDebug logs msg at debug level, see [Log].
func LogDebug(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.DebugLevel, msg, 1, kv...)
}

// LogInfo logs msg at info level, see [Log].
func LogInfo(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	Log(ctx, zapcore.InfoLevel, msg, 1, kv...)
}

// Log writes msg at level lvl to the global zap logger with kv as fields.
// It adds the trace and span IDs of a valid span in ctx and the code
// location of the caller skip frames above the caller of Log. Nothing is
// done if lvl is not enabled.
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

	if fr := goruntime.CallFrame(skip + 1); !fr.Zero() {
		attrs = append(attrs,
			semconv.CodeFunctionName(fr.Short()),
			semconv.CodeFilePath(path.Join(path.Base(path.Dir(fr.File)), path.Base(fr.File))),
			semconv.CodeLineNumber(fr.Line),
		)
	}

	zap.L().WithOptions(zap.WithCaller(false)).Log(lvl, msg, log.Attributes(attrs...)...)
}
