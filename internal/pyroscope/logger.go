package pyroscope

import (
	"fmt"

	"github.com/go-logr/logr"
)

type Logger struct {
	l logr.Logger
}

func NewLogger(l logr.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Infof(format string, a ...interface{}) {
	l.l.V(3).Info("[Info] " + fmt.Sprintf(format, a...))
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	l.l.V(4).Info("[Debug] " + fmt.Sprintf(format, a...))
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	l.l.V(0).Info("[Error] " + fmt.Sprintf(format, a...))
}
