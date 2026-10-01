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
	// Subscription is a service that manages the lifecycle of a goflux subscriber.
	// Create it with [NewSubscription].
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
	// SubscriptionOption configures a [Subscription] in [NewSubscription].
	SubscriptionOption[T any] func(*Subscription[T])
)

// NewSubscription returns a [Subscription] named name that subscribes
// subscriber to subject with handler. A nil l defaults to [log.Logger].
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

// Name returns the service name.
func (s *Subscription[T]) Name() string {
	return s.name
}

// Healthz returns [ErrServiceNotRunning] unless [Subscription.Start] is in
// progress.
func (s *Subscription[T]) Healthz() error {
	if !s.running.Load() {
		return ErrServiceNotRunning
	}

	return nil
}

// String returns a short description of the subject used in the readme.
func (s *Subscription[T]) String() string {
	return fmt.Sprintf("subject: `%s`", s.subject)
}

// Start subscribes to the subject and blocks until the subscription ends,
// returning the error from the subscriber.
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

// Close cancels the subscription context with cause [ErrServiceShutdown],
// closes the subscriber (logging a failure) and waits for the subscription to
// end. ctx is not used. Close must not be called before [Subscription.Start].
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
