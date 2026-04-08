package _chan

import (
	"context"
	"fmt"
	"sync"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/messaging"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Subscriber[T any] struct {
	bus     *Bus[T]
	bufSize int
	mu      sync.RWMutex
	ch      chan messaging.Message[T]
}

func NewSubscriber[T any](bus *Bus[T], bufSize int) (*Subscriber[T], error) {
	s := &Subscriber[T]{bus: bus, bufSize: bufSize}
	if _, err := messaging.RegisterLag(otel.GetMeterProvider(), "go_channel", s.Len); err != nil {
		return nil, fmt.Errorf("stream: register lag gauge: %w", err)
	}
	return s, nil
}

func (s *Subscriber[T]) Len() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ch == nil {
		return 0
	}
	return int64(len(s.ch))
}

func (s *Subscriber[T]) Subscribe(ctx context.Context, subject string, handler messaging.Handler[T]) error {
	ch := make(chan messaging.Message[T], s.bufSize)
	s.mu.Lock()
	s.ch = ch
	s.mu.Unlock()
	s.bus.subscribe(subject, ch)
	defer func() {
		s.bus.unsubscribe(subject, ch)
		s.mu.Lock()
		s.ch = nil
		s.mu.Unlock()
	}()

	l := log.Logger()

	for {
		select {
		case msg := <-ch:
			err := messaging.RecordProcess(ctx, subject, system, func(ctx context.Context) error {
				return handler(ctx, msg)
			})
			if err != nil {
				l.Error("stream: handler failed",
					zap.String("subject", subject),
					zap.Error(err),
				)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Subscriber[T]) Close() error { return nil }
