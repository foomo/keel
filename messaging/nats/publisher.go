// Package nats provides Publisher and Subscriber implementations backed
// by NATS core (not JetStream). For persistent / at-least-once delivery use
// the jetstream sub-package instead.
package nats

import (
	"context"
	"fmt"

	"github.com/foomo/keel/messaging"
	encodingx "github.com/foomo/keel/pkg/encoding"
	"github.com/nats-io/nats.go"
)

type Publisher[T any] struct {
	conn       *nats.Conn
	serializer encodingx.Codec[T]
}

func NewPublisher[T any](conn *nats.Conn, serializer encodingx.Codec[T]) *Publisher[T] {
	return &Publisher[T]{conn: conn, serializer: serializer}
}

func (p *Publisher[T]) Publish(ctx context.Context, subject string, v T) error {
	return messaging.RecordPublish(ctx, subject, system, func(ctx context.Context) error {
		b, err := p.serializer.Encode(v)
		if err != nil {
			return fmt.Errorf("nats publisher encode: %w", err)
		}
		return p.conn.Publish(subject, b)
	})
}

func (p *Publisher[T]) Close() error { return p.conn.Drain() }
