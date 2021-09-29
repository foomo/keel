package telemetry

import (
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type ErrorHandler struct {
	l *zap.Logger
}

func (h *ErrorHandler) Handle(err error) {
	if err != nil {
		log.WithError(h.l, err).Error("otel error")
	}
}

func (h *ErrorHandler) SetLogger(l *zap.Logger) {
	h.l = l
}