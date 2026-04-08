package messaging

import (
	"context"
)

// Filter decides whether a message should be forwarded.
// Returning false or a non-nil error drops the message and logs the reason.
type Filter[T any] func(ctx context.Context, msg Message[T]) (bool, error)
