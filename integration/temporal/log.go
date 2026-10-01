package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/activity"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// Error logs msg at error level on l with the given fields. The err argument
// is not included in the log entry.
func Error(l tlog.Logger, err error, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Error(msg)
}

// Info logs msg at info level on l with the given fields.
func Info(l tlog.Logger, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Info(msg)
}

// Warn logs msg at warn level on l with the given fields.
func Warn(l tlog.Logger, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Warn(msg)
}

// Debug logs msg at debug level on l with the given fields.
func Debug(l tlog.Logger, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Debug(msg)
}

// GetWorkflowLogger returns the workflow logger of ctx enriched with fields.
func GetWorkflowLogger(ctx workflow.Context, fields ...zap.Field) tlog.Logger {
	l := workflow.GetLogger(ctx)
	return LoggerWith(l, fields...)
}

// GetActivityLogger returns the activity logger of ctx enriched with fields.
func GetActivityLogger(ctx context.Context, fields ...zap.Field) tlog.Logger {
	l := activity.GetLogger(ctx)
	return LoggerWith(l, fields...)
}

// LoggerWith returns a logger that adds fields to every entry logged by l.
func LoggerWith(l tlog.Logger, fields ...zap.Field) tlog.Logger {
	v := make([]any, len(fields))
	for i, field := range fields {
		v[i] = field
	}

	return tlog.With(l, v...)
}
