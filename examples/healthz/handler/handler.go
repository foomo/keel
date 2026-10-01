package handler

import (
	"go.uber.org/zap"
)

// Handler is a health probe that logs its name and always reports healthy.
type Handler struct {
	l    *zap.Logger
	name string
}

// New returns a Handler logging with l under the given name.
func New(l *zap.Logger, name string) *Handler {
	return &Handler{l: l, name: name}
}

// Healthz logs the handler name and returns true.
func (h *Handler) Healthz() bool {
	h.l.Info(h.name)
	return true
}
