package _chan

import (
	"context"

	"github.com/foomo/keel/messaging"
)

type Publisher[T any] struct {
	bus *Bus[T]
}

func NewPublisher[T any](bus *Bus[T]) *Publisher[T] {
	return &Publisher[T]{bus: bus}
}

func (p *Publisher[T]) Publish(ctx context.Context, subject string, v T) error {
	msg := messaging.Message[T]{Subject: subject, Payload: v}
	return messaging.RecordPublish(ctx, subject, system, func(ctx context.Context) error {
		return p.bus.publish(ctx, subject, msg)
	})
}

func (p *Publisher[T]) Close() error { return nil }
