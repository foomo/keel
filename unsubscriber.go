package keel

import "context"

// Unsubscriber interface
type Unsubscriber interface {
	Unsubscribe()
}

// ErrorUnsubscriber interface
type ErrorUnsubscriber interface {
	Unsubscribe() error
}

// UnsubscriberWithContext interface
type UnsubscriberWithContext interface {
	Unsubscribe(ctx context.Context)
}

// ErrorUnsubscriberWithContext interface
type ErrorUnsubscriberWithContext interface {
	Unsubscribe(ctx context.Context) error
}
