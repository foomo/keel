package stream

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// MsgHandler handles a received NATS message. The logger is scoped to the
// stream the message was received on. A returned error is logged by the
// caller; it does not nak or redeliver the message.
type MsgHandler func(context.Context, *zap.Logger, *nats.Msg) error
