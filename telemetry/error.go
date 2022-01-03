package telemetry

import (
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type ErrorHandler struct {
	l *zap.Logger
}

func (h *ErrorHandler) Handle(err error) {
	l := log.WithError(h.l, err)
	if err != nil && err.Error() == "not implemented yet" {
		l.Warn("otel error")
	} else if err != nil {
		l.Error("otel error")
	}
}

func (h *ErrorHandler) SetLogger(l *zap.Logger) {
	h.l = l
}
