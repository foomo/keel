package keeltemporal

import (
	"fmt"

	"go.uber.org/zap"
)

type logger struct {
	l *zap.Logger
}

func NewLogger(l *zap.Logger) *logger {
	return &logger{
		l: l.WithOptions(zap.AddCallerSkip(3)),
	}
}

func (t *logger) Debug(msg string, keyvals ...interface{}) {
	t.l.Debug(msg, t.toFields(keyvals...)...)
}

func (t *logger) Info(msg string, keyvals ...interface{}) {
	t.l.Info(msg, t.toFields(keyvals...)...)
}

func (t *logger) Warn(msg string, keyvals ...interface{}) {
	t.l.Warn(msg, t.toFields(keyvals...)...)
}

func (t *logger) Error(msg string, keyvals ...interface{}) {
	t.l.Error(msg, t.toFields(keyvals...)...)
}

func (t *logger) toFields(keyvals ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keyvals); i++ {
		if value, ok := keyvals[i].(zap.Field); ok {
			fields = append(fields, value)
		} else if value, ok := keyvals[i].(string); ok && len(keyvals) > i+2 {
			fields = append(fields, zap.Any(value, keyvals[i+1]))
			i++
		} else if len(keyvals) > i+2 {
			fields = append(fields, zap.Any(fmt.Sprintf("%v", keyvals[i]), keyvals[i+1]))
			i++
		}
	}
	return fields
}
