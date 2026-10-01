package pyroscope

import (
	"fmt"

	"github.com/go-logr/logr"
)

// Logger adapts a [logr.Logger] to the pyroscope client logger interface.
type Logger struct {
	l logr.Logger
}

// NewLogger returns a [Logger] writing to l.
func NewLogger(l logr.Logger) *Logger {
	return &Logger{l: l}
}

// Infof logs a formatted message at verbosity 3.
func (l *Logger) Infof(format string, a ...any) {
	l.l.V(3).Info("[Info] " + fmt.Sprintf(format, a...))
}

// Debugf logs a formatted message at verbosity 4.
func (l *Logger) Debugf(format string, a ...any) {
	l.l.V(4).Info("[Debug] " + fmt.Sprintf(format, a...))
}

// Errorf logs a formatted message at verbosity 0 as an info entry.
func (l *Logger) Errorf(format string, a ...any) {
	l.l.V(0).Info("[Error] " + fmt.Sprintf(format, a...))
}
