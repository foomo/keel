package interfaces

import (
	"context"
)

type closer struct {
	handle func(context.Context) error
}

func NewCloserFn(handle func(context.Context) error) closer {
	return closer{
		handle: handle,
	}
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
