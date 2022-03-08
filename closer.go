package keel

import "context"

// Closer interface
type Closer interface {
	Close()
}

// CloserFn interface
type CloserFn func()

// ErrorCloserFn interface
type ErrorCloserFn func() error

// ErrorCloser interface
type ErrorCloser interface {
	Close() error
}

// CloserWithContext interface
type CloserWithContext interface {
	Close(ctx context.Context)
}

// CloserWithContextFn interface
type CloserWithContextFn func(ctx context.Context)

// ErrorCloserWithContext interface
type ErrorCloserWithContext interface {
	Close(ctx context.Context) error
}

// ErrorCloserWithContextFn interface
type ErrorCloserWithContextFn func(ctx context.Context) error

// Shutdowner interface
type Shutdowner interface {
	Shutdown()
}

// ErrorShutdowner interface
type ErrorShutdowner interface {
	Shutdown() error
}

// ShutdownerWithContext interface
type ShutdownerWithContext interface {
	Shutdown(ctx context.Context)
}

// ErrorShutdownerWithContext interface
type ErrorShutdownerWithContext interface {
	Shutdown(ctx context.Context) error
}

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
