package interfaces

import "context"

// Stopper is implemented by any value that exposes a Stop method without a
// return value. It is the simplest of the four Stop shapes; see also
// [ErrorStopper], [StopperWithContext] and [ErrorStopperWithContext].
type Stopper interface {
	Stop()
}

// IsStopper reports whether v implements [Stopper] and returns the
// asserted value. The boolean is false when v does not implement
// [Stopper].
func IsStopper(v any) (Stopper, bool) {
	return Is[Stopper](v)
}

// ErrorStopper is like [Stopper] but its Stop method returns an error.
type ErrorStopper interface {
	Stop() error
}

// IsErrorStopper reports whether v implements [ErrorStopper] and returns
// the asserted value. The boolean is false when v does not implement
// [ErrorStopper].
func IsErrorStopper(v any) (ErrorStopper, bool) {
	return Is[ErrorStopper](v)
}

// StopperWithContext is like [Stopper] but its Stop method accepts a
// [context.Context] so the caller can bound the operation.
type StopperWithContext interface {
	Stop(ctx context.Context)
}

// IsStopperWithContext reports whether v implements [StopperWithContext]
// and returns the asserted value. The boolean is false when v does not
// implement [StopperWithContext].
func IsStopperWithContext(v any) (StopperWithContext, bool) {
	return Is[StopperWithContext](v)
}

// ErrorStopperWithContext combines [ErrorStopper] and [StopperWithContext]:
// its Stop method accepts a [context.Context] and returns an error.
type ErrorStopperWithContext interface {
	Stop(ctx context.Context) error
}

// IsErrorStopperWithContext reports whether v implements
// [ErrorStopperWithContext] and returns the asserted value. The boolean is
// false when v does not implement [ErrorStopperWithContext].
func IsErrorStopperWithContext(v any) (ErrorStopperWithContext, bool) {
	return Is[ErrorStopperWithContext](v)
}
