package handler

import (
	"context"

	"go.uber.org/zap"
)

type Handler struct {
	l *zap.Logger
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(ctx context.Context) error {
	h.l.Info("ping")
	return nil
}
