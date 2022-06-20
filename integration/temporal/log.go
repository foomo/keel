package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/activity"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func Error(l tlog.Logger, err error, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Error(msg)
}

func Info(l tlog.Logger, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Info(msg)
}

func Warn(l tlog.Logger, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Warn(msg)
}

func Debug(l tlog.Logger, msg string, fields ...zap.Field) {
	LoggerWith(l, fields...).Debug(msg)
}

func GetWorkflowLogger(ctx workflow.Context, fields ...zap.Field) tlog.Logger {
	l := workflow.GetLogger(ctx)
	return LoggerWith(l, fields...)
}

func GetActivityLogger(ctx context.Context, fields ...zap.Field) tlog.Logger {
	l := activity.GetLogger(ctx)
	return LoggerWith(l, fields...)
}

func LoggerWith(l tlog.Logger, fields ...zap.Field) tlog.Logger {
	v := make([]interface{}, len(fields))
	for i, field := range fields {
		v[i] = field
	}
	return tlog.With(l, v...)
}
