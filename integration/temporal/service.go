package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// service runs a Temporal worker as a keel service.
type service struct {
	l    *zap.Logger
	w    worker.Worker
	name string
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewService returns a keel service named name that starts w on Start and
// stops it on Close. If l is nil, the keel default logger is used.
func NewService(l *zap.Logger, name string, w worker.Worker) *service {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithHTTPServerName(l, name)

	return &service{l: l, name: name, w: w}
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// Name returns the service name.
func (s *service) Name() string {
	return s.name
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Start starts the worker without blocking and returns its start error.
func (s *service) Start(ctx context.Context) error {
	s.l.Info("starting temporal worker")
	return s.w.Start()
}

// Close stops the worker. It always returns nil.
func (s *service) Close(ctx context.Context) error {
	s.l.Info("stopping temporal worker")
	s.w.Stop()

	return nil
}
