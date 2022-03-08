package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/worker"
)

type test struct {
	w    worker.Worker
	name string
}

func NewTestService(name string, w worker.Worker) *test {
	return &test{name: name, w: w}
}

func (s *test) URL() string {
	return ""
}

func (s *test) Name() string {
	return s.name
}

func (s *test) Start(ctx context.Context) error {
	_ = s.w.Start()
	return nil
}

func (s *test) Close(ctx context.Context) error {
	s.w.Stop()
	return nil
}
