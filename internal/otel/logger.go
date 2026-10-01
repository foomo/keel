package otel

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/log"
)

// Logger is a [logr.LogSink] writing to a zap logger, used as the
// OpenTelemetry internal logger.
type Logger struct {
	l *zap.Logger
}

// NewLogger returns a [Logger] writing to l.
func NewLogger(l *zap.Logger) Logger {
	return Logger{l: l}
}

// Init implements [logr.LogSink] and does nothing.
func (l Logger) Init(info logr.RuntimeInfo) {
}

// Enabled reports whether the logr verbosity level, mapped to the negated
// zap level, is enabled by the global keel log level.
func (l Logger) Enabled(level int) bool {
	return log.AtomicLevel().Enabled(zapcore.Level(-1 * level)) //nolint:gosec
}

// Info logs msg at zap info level regardless of level.
func (l Logger) Info(level int, msg string, keysAndValues ...any) {
	l.l.Info(msg, l.fields(keysAndValues)...)
}

// Error logs msg at zap error level. err is not included in the entry.
func (l Logger) Error(err error, msg string, keysAndValues ...any) {
	l.l.Error(msg, l.fields(keysAndValues)...)
}

// WithValues returns a sink with keysAndValues added as fields.
func (l Logger) WithValues(keysAndValues ...any) logr.LogSink {
	return NewLogger(l.l.With(l.fields(keysAndValues)...))
}

// WithName returns a sink with name appended to the logger name.
func (l Logger) WithName(name string) logr.LogSink {
	return NewLogger(l.l.Named(name))
}

// fields converts alternating keys and values to zap fields. It expects an
// even number of elements.
func (l Logger) fields(keysAndValues []any) []zap.Field {
	ret := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		ret = append(ret, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
	}

	return ret
}
