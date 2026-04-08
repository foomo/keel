package messaging

import (
	"context"
)

// DeadLetterFunc receives messages that could not be mapped or published after
// all retries are exhausted. Use it to log, alert, or forward to a DLQ.
type DeadLetterFunc[T any] func(ctx context.Context, msg Message[T], err error)
