package messaging

import (
	"context"
)

// Handler is the callback signature used by Subscriber.Subscribe.
// Returning a non-nil error signals the subscriber to nack / requeue the
// message (behavior is implementation-specific).
type Handler[T any] func(ctx context.Context, msg Message[T]) error
