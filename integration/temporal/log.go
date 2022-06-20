package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/activity"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

func Error(ctx workflow.Context, err error, msg string, fields ...zap.Field) {
	keyvals := make([]interface{}, 0, len(fields)+1)
	keyvals = append(keyvals, log.FError(err))
	for _, field := range fields {
		keyvals = append(keyvals, field)
	}
	workflow.GetLogger(ctx).Error(msg, keyvals...)
}

func Info(ctx workflow.Context, msg string, fields ...zap.Field) {
	keyvals := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		keyvals = append(keyvals, field)
	}
	workflow.GetLogger(ctx).Info(msg, keyvals...)
}

func Debug(ctx workflow.Context, msg string, fields ...zap.Field) {
	keyvals := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		keyvals = append(keyvals, field)
	}
	workflow.GetLogger(ctx).Debug(msg, keyvals...)
}

func GetWorkflowLogger(ctx workflow.Context, fields ...zap.Field) tlog.Logger {
	l := workflow.GetLogger(ctx)
	return LoggerWith(l, fields...)
}

func GetActivityLogger(ctx context.Context, fields ...zap.Field) tlog.Logger {
	l := activity.GetLogger(ctx)
	return LoggerWith(l, fields...)
}

func LoggerWith(logger tlog.Logger, fields ...zap.Field) tlog.Logger {
	v := make([]interface{}, len(fields))
	for i, field := range fields {
		v[i] = field
	}
	return tlog.With(logger, v...)
}
