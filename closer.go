package keel

import "context"

type closer struct {
	handle func(context.Context) error
}

func NewCloserFn(handle func(context.Context) error) closer {
	return closer{
		handle: handle,
	}
}

func (h healther) Close(ctx context.Context) error {
	return h.handle(ctx)
}

// Closer interface
type Closer interface {
	Close()
}

// ErrorCloser interface
type ErrorCloser interface {
	Close() error
}

// CloserWithContext interface
type CloserWithContext interface {
	Close(ctx context.Context)
}

// ErrorCloserWithContext interface
type ErrorCloserWithContext interface {
	Close(ctx context.Context) error
}
