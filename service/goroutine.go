package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/foomo/keel/log"
	keelsemconv "github.com/foomo/keel/semconv"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type (
	// GoRoutine is a service that runs a [GoRoutineFn] in one or more
	// goroutines until it returns or the service is closed. Create it with
	// [NewGoRoutine].
	GoRoutine struct {
		running    atomic.Bool
		handler    GoRoutineFn
		cancel     context.CancelCauseFunc
		cancelLock sync.Mutex
		parallel   int
		name       string
		wg         errgroup.Group
		l          *zap.Logger
	}
	// GoRoutineOption configures a [GoRoutine] in [NewGoRoutine].
	GoRoutineOption func(*GoRoutine)
	// GoRoutineFn is the function run by a [GoRoutine]. ctx is canceled with
	// cause [ErrServiceShutdown] when the service is closed; l is enriched with
	// the service name and goroutine instance.
	GoRoutineFn func(ctx context.Context, l *zap.Logger) error
)

// NewGoRoutine returns a [GoRoutine] named name that runs handler. A nil l
// defaults to [log.Logger]. By default handler runs in a single goroutine.
func NewGoRoutine(l *zap.Logger, name string, handler GoRoutineFn, opts ...GoRoutineOption) *GoRoutine {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithAttributes(l,
		keelsemconv.KeelServiceType(keelsemconv.KeelServiceTypeGoRoutine),
		keelsemconv.KeelServiceName(name),
	)

	inst := &GoRoutine{
		handler:  handler,
		name:     name,
		parallel: 1,
		l:        l,
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// GoRoutineWithParallel sets the number of goroutines running the handler
// concurrently. Defaults to 1.
func GoRoutineWithParallel(v int) GoRoutineOption {
	return func(o *GoRoutine) {
		o.parallel = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Name returns the service name.
func (s *GoRoutine) Name() string {
	return s.name
}

// Healthz returns [ErrServiceNotRunning] unless [GoRoutine.Start] is in progress.
func (s *GoRoutine) Healthz() error {
	if !s.running.Load() {
		return ErrServiceNotRunning
	}

	return nil
}

// String returns a short description used in the readme.
func (s *GoRoutine) String() string {
	return fmt.Sprintf("parallel: `%d`", s.parallel)
}

// Start runs the handler in the configured number of goroutines and blocks
// until all of them have returned. It returns the first non-nil handler error.
func (s *GoRoutine) Start(ctx context.Context) error {
	s.l.Info("starting keel service")

	ctx, cancel := context.WithCancelCause(ctx)

	s.cancelLock.Lock()
	s.cancel = cancel
	s.cancelLock.Unlock()

	for i := range s.parallel {
		l := log.WithAttributes(s.l, keelsemconv.KeelServiceInst(i))
		s.wg.Go(func() error {
			return s.handler(ctx, l)
		})
	}

	s.running.Store(true)

	defer func() {
		s.running.Store(false)
	}()

	return s.wg.Wait()
}

// Close cancels the handler context with cause [ErrServiceShutdown] and waits
// for all goroutines to return, returning the first non-nil handler error. ctx
// is not used. Close must not be called before [GoRoutine.Start].
func (s *GoRoutine) Close(ctx context.Context) error {
	s.l.Info("stopping keel service")
	s.cancelLock.Lock()
	s.cancel(ErrServiceShutdown)
	s.cancelLock.Unlock()

	return s.wg.Wait()
}
