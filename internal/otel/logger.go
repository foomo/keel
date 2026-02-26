package otel

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/log"
)

type Logger struct {
	l *zap.Logger
}

func NewLogger(l *zap.Logger) Logger {
	return Logger{l: l}
}

func (l Logger) Init(info logr.RuntimeInfo) {
}

func (l Logger) Enabled(level int) bool {
	return log.AtomicLevel().Enabled(zapcore.Level(-1 * level)) //nolint:gosec
}

func (l Logger) Info(level int, msg string, keysAndValues ...any) {
	l.l.Info(msg, l.fields(keysAndValues)...)
}

func (l Logger) Error(err error, msg string, keysAndValues ...any) {
	l.l.Error(msg, l.fields(keysAndValues)...)
}

func (l Logger) WithValues(keysAndValues ...any) logr.LogSink {
	return NewLogger(l.l.With(l.fields(keysAndValues)...))
}

func (l Logger) WithName(name string) logr.LogSink {
	return NewLogger(l.l.Named(name))
}

func (l Logger) fields(keysAndValues []any) []zap.Field {
	ret := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		ret = append(ret, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
	}

	return ret
}
