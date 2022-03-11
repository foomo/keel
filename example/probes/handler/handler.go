package handler

import (
	"go.uber.org/zap"
)

type Handler struct {
	l    *zap.Logger
	name string
}

func New(l *zap.Logger, name string) *Handler {
	return &Handler{l: l, name: name}
}

func (h *Handler) Healthz() bool {
	h.l.Info(h.name)
	return true
}
