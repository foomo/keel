package interfaces

import "context"

// Shutdowner is implemented by any value that exposes a Shutdown method
// without a return value. It is the simplest of the four Shutdown shapes;
// see also [ErrorShutdowner], [ShutdownerWithContext] and
// [ErrorShutdownerWithContext].
type Shutdowner interface {
	Shutdown()
}

// IsShutdowner reports whether v implements [Shutdowner] and returns the
// asserted value. The boolean is false when v does not implement
// [Shutdowner].
func IsShutdowner(v any) (Shutdowner, bool) {
	return Is[Shutdowner](v)
}

// ErrorShutdowner is like [Shutdowner] but its Shutdown method returns an
// error.
type ErrorShutdowner interface {
	Shutdown() error
}

// IsErrorShutdowner reports whether v implements [ErrorShutdowner] and
// returns the asserted value. The boolean is false when v does not
// implement [ErrorShutdowner].
func IsErrorShutdowner(v any) (ErrorShutdowner, bool) {
	return Is[ErrorShutdowner](v)
}

// ShutdownerWithContext is like [Shutdowner] but its Shutdown method
// accepts a [context.Context] so the caller can bound the operation.
type ShutdownerWithContext interface {
	Shutdown(ctx context.Context)
}

// IsShutdownerWithContext reports whether v implements
// [ShutdownerWithContext] and returns the asserted value. The boolean is
// false when v does not implement [ShutdownerWithContext].
func IsShutdownerWithContext(v any) (ShutdownerWithContext, bool) {
	return Is[ShutdownerWithContext](v)
}

// ErrorShutdownerWithContext combines [ErrorShutdowner] and
// [ShutdownerWithContext]: its Shutdown method accepts a
// [context.Context] and returns an error.
type ErrorShutdownerWithContext interface {
	Shutdown(ctx context.Context) error
}

// IsErrorShutdownerWithContext reports whether v implements
// [ErrorShutdownerWithContext] and returns the asserted value. The boolean
// is false when v does not implement [ErrorShutdownerWithContext].
func IsErrorShutdownerWithContext(v any) (ErrorShutdownerWithContext, bool) {
	return Is[ErrorShutdownerWithContext](v)
}
