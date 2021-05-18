package keel

import "context"

// Closer interface
type Closer interface {
	Close(ctx context.Context) error
}

// Shutdowner interface
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}
