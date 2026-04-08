// Package nats provides Publisher and Subscriber implementations backed
// by NATS core (not JetStream). For persistent / at-least-once delivery use
// the jetstream sub-package instead.
package nats

import (
	"context"
	"fmt"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/messaging"
	encodingx "github.com/foomo/keel/pkg/encoding"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Subscriber[T any] struct {
	conn       *nats.Conn
	serializer encodingx.Codec[T]
}

func NewSubscriber[T any](conn *nats.Conn, serializer encodingx.Codec[T]) *Subscriber[T] {
	return &Subscriber[T]{conn: conn, serializer: serializer}
}

func (s *Subscriber[T]) Subscribe(ctx context.Context, subject string, handler messaging.Handler[T]) error {
	l := log.Logger()

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		var v T
		if err := s.serializer.Decode(msg.Data, &v); err != nil {
			l.Warn("nats: decode failed, dropping message",
				zap.String("subject", msg.Subject),
				zap.Error(err),
			)
			return
		}
		m := messaging.Message[T]{Subject: msg.Subject, Payload: v}
		if err := messaging.RecordProcess(ctx, subject, system, func(ctx context.Context) error {
			return handler(ctx, m)
		}); err != nil {
			l.Error("nats: handler failed",
				zap.String("subject", msg.Subject),
				zap.Error(err),
			)
		}
	})
	if err != nil {
		return fmt.Errorf("nats subscriber: %w", err)
	}
	<-ctx.Done()
	return sub.Unsubscribe()
}

func (s *Subscriber[T]) Close() error { return s.conn.Drain() }
