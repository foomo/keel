package keeltemporal

import (
	"context"

	"go.temporal.io/sdk/worker"
)

// test runs a Temporal worker as a keel service in tests.
type test struct {
	w    worker.Worker
	name string
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// NewTestService returns a keel service named name that starts w on Start and
// stops it on Close. Unlike [NewService] it does not log and ignores worker
// start errors.
func NewTestService(name string, w worker.Worker) *test {
	return &test{name: name, w: w}
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// URL returns an empty string.
func (s *test) URL() string {
	return ""
}

// Name returns the service name.
func (s *test) Name() string {
	return s.name
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Start starts the worker without blocking. It always returns nil.
func (s *test) Start(ctx context.Context) error {
	_ = s.w.Start()
	return nil
}

// Close stops the worker. It always returns nil.
func (s *test) Close(ctx context.Context) error {
	s.w.Stop()
	return nil
}
