package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/worker"
)

type service struct {
	w    worker.Worker
	name string
}

func NewService(name string, w worker.Worker) *service {
	return &service{name: name, w: w}
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Start(ctx context.Context) error {
	return s.w.Start()
}

func (s *service) Close(ctx context.Context) error {
	s.w.Stop()
	return nil
}
