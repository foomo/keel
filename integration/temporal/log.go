package keeltemporal

import (
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
