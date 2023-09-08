package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/foomo/keel/log"
)

// GoRoutine struct
type (
	GoRoutine struct {
		running  atomic.Bool
		handler  GoRoutineFn
		cancel   context.CancelCauseFunc
		parallel int
		name     string
		wg       errgroup.Group
		l        *zap.Logger
	}
	GoRoutineOption func(*GoRoutine)
	GoRoutineFn     func(ctx context.Context, l *zap.Logger) error
)

func NewGoRoutine(l *zap.Logger, name string, handler GoRoutineFn) *GoRoutine {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithAttributes(l,
		log.KeelServiceTypeKey.String("goroutine"),
		log.KeelServiceNameKey.String(name),
	)

	return &GoRoutine{
		handler:  handler,
		name:     name,
		parallel: 1,
		l:        l,
	}
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
	s.cancel = cancel
	for i := 0; i < s.parallel; i++ {
		i := i
		l := log.WithAttributes(s.l, log.KeelServiceInstKey.Int(i))
		s.wg.Go(func() error {
			return s.handler(ctx, l)
		})
	}
	return s.wg.Wait()
}

func (s *GoRoutine) Close(ctx context.Context) error {
	s.l.Info("stopping keel service")
	s.cancel(ErrServiceShutdown)
	return s.wg.Wait()
}
