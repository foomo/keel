package stream

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type MsgHandler func(context.Context, *zap.Logger, *nats.Msg) error
