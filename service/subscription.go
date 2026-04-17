package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/foomo/goflux"
	"github.com/foomo/keel/log"
	keelsemconv "github.com/foomo/keel/semconv"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type (
	// Subscription is a Service that manages the lifecycle of a goflux subscriber.
	Subscription[T any] struct {
		running    atomic.Bool
		subscriber goflux.Subscriber[T]
		subject    string
		handler    goflux.Handler[T]
		cancel     context.CancelCauseFunc
		cancelLock sync.Mutex
		name       string
		wg         errgroup.Group
		l          *zap.Logger
	}
	SubscriptionOption[T any] func(*Subscription[T])
)

// NewSubscription creates a new Subscription service that runs the given
// subscriber as a managed keel service.
func NewSubscription[T any](
	l *zap.Logger,
	name string,
	subscriber goflux.Subscriber[T],
	subject string,
	handler goflux.Handler[T],
	opts ...SubscriptionOption[T],
) *Subscription[T] {
	if l == nil {
		l = log.Logger()
	}

	l = log.WithAttributes(l,
		keelsemconv.KeelServiceType("sub"),
		keelsemconv.KeelServiceName(name),
	)

	inst := &Subscription[T]{
		subscriber: subscriber,
		subject:    subject,
		handler:    handler,
		name:       name,
		l:          l,
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *Subscription[T]) Name() string {
	return s.name
}

func (s *Subscription[T]) Healthz() error {
	if !s.running.Load() {
		return ErrServiceNotRunning
	}

	return nil
}

func (s *Subscription[T]) String() string {
	return fmt.Sprintf("subject: `%s`", s.subject)
}

func (s *Subscription[T]) Start(ctx context.Context) error {
	s.l.Info("starting keel service")

	ctx, cancel := context.WithCancelCause(ctx)

	s.cancelLock.Lock()
	s.cancel = cancel
	s.cancelLock.Unlock()

	l := s.l
	s.wg.Go(func() error {
		l.Info("subscribing", zap.String("subject", s.subject))
		return s.subscriber.Subscribe(ctx, s.subject, s.handler)
	})

	s.running.Store(true)

	defer func() {
		s.running.Store(false)
	}()

	return s.wg.Wait()
}

func (s *Subscription[T]) Close(ctx context.Context) error {
	s.l.Info("stopping keel service")

	s.cancelLock.Lock()
	s.cancel(ErrServiceShutdown)
	s.cancelLock.Unlock()

	if err := s.subscriber.Close(); err != nil {
		s.l.Warn("subscriber close failed", zap.Error(err))
	}

	return s.wg.Wait()
}
