package interfaces

import (
	"context"
)

// Closer is implemented by any value that exposes a Close method without a
// return value. It is the simplest of the four Close shapes; see also
// [ErrorCloser], [CloserWithContext] and [ErrorCloserWithContext].
type Closer interface {
	Close()
}

// IsCloser reports whether v implements [Closer] and returns the asserted
// value. The boolean is false when v does not implement [Closer].
func IsCloser(v any) (Closer, bool) {
	return Is[Closer](v)
}

// ErrorCloser is like [Closer] but its Close method returns an error.
type ErrorCloser interface {
	Close() error
}

// IsErrorCloser reports whether v implements [ErrorCloser] and returns the
// asserted value. The boolean is false when v does not implement
// [ErrorCloser].
func IsErrorCloser(v any) (ErrorCloser, bool) {
	return Is[ErrorCloser](v)
}

// CloserWithContext is like [Closer] but its Close method accepts a
// [context.Context] so the caller can bound the close operation.
type CloserWithContext interface {
	Close(ctx context.Context)
}

// IsCloserWithContext reports whether v implements [CloserWithContext] and
// returns the asserted value. The boolean is false when v does not
// implement [CloserWithContext].
func IsCloserWithContext(v any) (CloserWithContext, bool) {
	return Is[CloserWithContext](v)
}

// ErrorCloserWithContext combines [ErrorCloser] and [CloserWithContext]:
// its Close method accepts a [context.Context] and returns an error.
type ErrorCloserWithContext interface {
	Close(ctx context.Context) error
}

// IsErrorCloserWithContext reports whether v implements
// [ErrorCloserWithContext] and returns the asserted value. The boolean is
// false when v does not implement [ErrorCloserWithContext].
func IsErrorCloserWithContext(v any) (ErrorCloserWithContext, bool) {
	return Is[ErrorCloserWithContext](v)
}
