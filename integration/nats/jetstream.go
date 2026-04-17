package nats

import (
	"context"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/opentelemetry-go/semconv/natsconv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// NewJetStream creates a JetStream context with production-sensible async
// publish defaults and an instrumented async-error handler.
func NewJetStream(s *keel.Server, nc *nats.Conn, opts ...jetstream.JetStreamOpt) (jetstream.JetStream, error) {
	l := s.Logger().Named("jetstream")
	m := s.Meter()

	asyncErrors, err := natsconv.NewClientAsyncErrors(m)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Prepend defaults so the caller can still override them via opts...
	opts = append([]jetstream.JetStreamOpt{
		jetstream.WithDefaultTimeout(5 * time.Second),
		jetstream.WithPublishAsyncMaxPending(256),
		jetstream.WithPublishAsyncTimeout(10 * time.Second),
		jetstream.WithPublishAsyncErrHandler(func(_ jetstream.JetStream, msg *nats.Msg, err error) {
			asyncErrors.Add(ctx, 1, natsconv.AsyncErrorKindOther,
				asyncErrors.AttrSubject(msg.Subject),
			)
			l.Error("publish async error",
				zap.Error(err),
				zap.String("subject", msg.Subject),
			)
		}),
	}, opts...)

	return jetstream.New(nc, opts...)
}
