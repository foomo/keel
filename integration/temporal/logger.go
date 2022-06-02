package keeltemporal

import (
	"fmt"
	"strings"

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
	t.l.Debug(msg, t.fields(keyvals...)...)
}

func (t *logger) Info(msg string, keyvals ...interface{}) {
	// TODO check with temporal why errors are being logged as info!
	for _, keyval := range keyvals {
		if keyval == "Error" {
			t.l.Error(msg, t.fields(keyvals...)...)
			return
		}
	}
	t.l.Info(msg, t.fields(keyvals...)...)
}

func (t *logger) Warn(msg string, keyvals ...interface{}) {
	t.l.Warn(msg, t.fields(keyvals...)...)
}

func (t *logger) Error(msg string, keyvals ...interface{}) {
	t.l.Error(msg, t.fields(keyvals...)...)
}

func (t *logger) With(keyvals ...interface{}) *logger {
	return NewLogger(t.l.With(t.fields(keyvals...)...))
}

func (t *logger) fields(keyvals ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keyvals); i++ {
		if value, ok := keyvals[i].(zap.Field); ok {
			fields = append(fields, value)
		} else if value, ok := keyvals[i].(string); ok && len(keyvals) > i+2 {
			fields = append(fields, zap.Any(strings.ToLower(value), keyvals[i+1]))
			i++
		} else if len(keyvals) > i+1 {
			fields = append(fields, zap.Any(strings.ToLower(fmt.Sprintf("%v", keyvals[i])), keyvals[i+1]))
			i++
		} else {
			fields = append(fields, zap.Any("undefined", keyvals[i]))
		}
	}
	return fields
}
