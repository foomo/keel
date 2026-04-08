package messaging

import (
	"context"
)

// Publisher sends encoded messages to a subject/topic.
type Publisher[T any] interface {
	// Publish serializes v via the bound Codec and delivers it to the subject.
	Publish(ctx context.Context, subject string, v T) error
	// Close releases any underlying connections.
	Close() error
}
