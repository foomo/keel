package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	"github.com/pkg/errors"
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

type toggleableService struct {
	l         *zap.Logger
	newWorker func() worker.Worker
	w         worker.Worker
	name      string
}

func NewToggleableService(l *zap.Logger, name string, newWorker func() worker.Worker) *toggleableService {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithHTTPServerName(l, name)
	return &toggleableService{l: l, name: name, newWorker: newWorker}
}

func (s *toggleableService) Name() string {
	return s.name
}

func (s *toggleableService) Start(ctx context.Context) error {
	if s.w != nil {
		// to prevent memory leaks, block starting a worker if it has already been started
		return errors.New("can not start new worker while another one is running")
	}
	s.l.Info("creating and starting temporal worker")
	s.w = s.newWorker()
	return s.w.Start()
}

func (s *toggleableService) Close(ctx context.Context) error {
	if s.w != nil {
		s.l.Info("stopping temporal worker")
		s.w.Stop()
		s.w = nil
	}
	return nil
}
