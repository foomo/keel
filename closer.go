package keel

import "context"

// Closer interface
type Closer interface {
	Close() error
}

// CloserWithContext interface
type CloserWithContext interface {
	Close(ctx context.Context) error
}

// Shutdowner interface
type Shutdowner interface {
	Shutdown() error
}

// ShutdownerWithContext interface
type ShutdownerWithContext interface {
	Shutdown(ctx context.Context) error
}

// Unsubscriber interface
type Unsubscriber interface {
	Unsubscribe() error
}

// UnsubscriberWithContext interface
type UnsubscriberWithContext interface {
	Unsubscribe(ctx context.Context) error
}
