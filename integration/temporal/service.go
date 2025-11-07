package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type service struct {
	l    *zap.Logger
	w    worker.Worker
	name string
}

func NewService(l *zap.Logger, name string, w worker.Worker) *service {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithHTTPServerName(l, name)

	return &service{l: l, name: name, w: w}
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Start(ctx context.Context) error {
	s.l.Info("starting temporal worker")
	return s.w.Start()
}

func (s *service) Close(ctx context.Context) error {
	s.l.Info("stopping temporal worker")
	s.w.Stop()

	return nil
}
