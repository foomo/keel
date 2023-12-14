package interfaces

import (
	"context"
)

type CloseHandler struct {
	Value func(ctx context.Context) error
}

func (r CloseHandler) Close(ctx context.Context) error {
	return r.Value(ctx)
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

func CloseFunc(v func(ctx context.Context) error) CloseHandler {
	return CloseHandler{
		Value: v,
	}
}
