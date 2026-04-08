package messaging

import (
	"context"
)

// MapFunc transforms a Message[T] into a Message[U].
// A non-nil error drops the message and routes it to the DeadLetterFunc if set.
type MapFunc[T, U any] func(ctx context.Context, msg Message[T]) (Message[U], error)
