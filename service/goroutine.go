package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/foomo/keel/log"
)

// GoRoutine struct
type (
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
	GoRoutineOption func(*GoRoutine)
	GoRoutineFn     func(ctx context.Context, l *zap.Logger) error
)

func NewGoRoutine(l *zap.Logger, name string, handler GoRoutineFn, opts ...GoRoutineOption) *GoRoutine {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithAttributes(l,
		log.KeelServiceTypeKey.String("goroutine"),
		log.KeelServiceNameKey.String(name),
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

func GoRoutineWithPralllel(v int) GoRoutineOption {
	return func(o *GoRoutine) {
		o.parallel = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *GoRoutine) Name() string {
	return s.name
}

func (s *GoRoutine) Healthz() error {
	if !s.running.Load() {
		return ErrServiceNotRunning
	}

	return nil
}

func (s *GoRoutine) String() string {
	return fmt.Sprintf("parallel: `%d`", s.parallel)
}

func (s *GoRoutine) Start(ctx context.Context) error {
	s.l.Info("starting keel service")

	ctx, cancel := context.WithCancelCause(ctx)

	s.cancelLock.Lock()
	s.cancel = cancel
	s.cancelLock.Unlock()

	for i := range s.parallel {
		l := log.WithAttributes(s.l, log.KeelServiceInstKey.Int(i))
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

func (s *GoRoutine) Close(ctx context.Context) error {
	s.l.Info("stopping keel service")
	s.cancelLock.Lock()
	s.cancel(ErrServiceShutdown)
	s.cancelLock.Unlock()

	return s.wg.Wait()
}
