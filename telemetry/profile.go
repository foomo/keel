package telemetry

import (
	"fmt"

	"go.uber.org/zap"
)

type PyroscopeLogger struct {
	l *zap.Logger
}

func NewPyroscopeLogger(l *zap.Logger) *PyroscopeLogger {
	return &PyroscopeLogger{l: l.Named("pyroscope")}
}

func (l *PyroscopeLogger) Infof(format string, a ...interface{}) {
	l.l.Info(fmt.Sprintf(format, a...))
}

func (l *PyroscopeLogger) Debugf(format string, a ...interface{}) {
	l.l.Debug(fmt.Sprintf(format, a...))
}

func (l *PyroscopeLogger) Errorf(format string, a ...interface{}) {
	l.l.Error(fmt.Sprintf(format, a...))
}
