package telemetry

import (
	"context"

	"github.com/foomo/keel/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Log(ctx context.Context) *zap.Logger {
	span := trace.SpanContextFromContext(ctx)
	var fields []zapcore.Field
	if span.IsValid() {
		fields = append(fields,
			zap.String("trace_id", span.TraceID().String()),
			zap.String("span_id", span.SpanID().String()),
		)
	}
	return log.Logger().With(fields...)
}
