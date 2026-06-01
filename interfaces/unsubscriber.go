package interfaces

import "context"

// Unsubscriber is implemented by any value that exposes an Unsubscribe
// method without a return value. It is the simplest of the four
// Unsubscribe shapes; see also [ErrorUnsubscriber],
// [UnsubscriberWithContext] and [ErrorUnsubscriberWithContext].
type Unsubscriber interface {
	Unsubscribe()
}

// IsUnsubscriber reports whether v implements [Unsubscriber] and returns
// the asserted value. The boolean is false when v does not implement
// [Unsubscriber].
func IsUnsubscriber(v any) (Unsubscriber, bool) {
	return Is[Unsubscriber](v)
}

// ErrorUnsubscriber is like [Unsubscriber] but its Unsubscribe method
// returns an error.
type ErrorUnsubscriber interface {
	Unsubscribe() error
}

// IsErrorUnsubscriber reports whether v implements [ErrorUnsubscriber] and
// returns the asserted value. The boolean is false when v does not
// implement [ErrorUnsubscriber].
func IsErrorUnsubscriber(v any) (ErrorUnsubscriber, bool) {
	return Is[ErrorUnsubscriber](v)
}

// UnsubscriberWithContext is like [Unsubscriber] but its Unsubscribe
// method accepts a [context.Context] so the caller can bound the
// operation.
type UnsubscriberWithContext interface {
	Unsubscribe(ctx context.Context)
}

// IsUnsubscriberWithContext reports whether v implements
// [UnsubscriberWithContext] and returns the asserted value. The boolean is
// false when v does not implement [UnsubscriberWithContext].
func IsUnsubscriberWithContext(v any) (UnsubscriberWithContext, bool) {
	return Is[UnsubscriberWithContext](v)
}

// ErrorUnsubscriberWithContext combines [ErrorUnsubscriber] and
// [UnsubscriberWithContext]: its Unsubscribe method accepts a
// [context.Context] and returns an error.
type ErrorUnsubscriberWithContext interface {
	Unsubscribe(ctx context.Context) error
}

// IsErrorUnsubscriberWithContext reports whether v implements
// [ErrorUnsubscriberWithContext] and returns the asserted value. The
// boolean is false when v does not implement
// [ErrorUnsubscriberWithContext].
func IsErrorUnsubscriberWithContext(v any) (ErrorUnsubscriberWithContext, bool) {
	return Is[ErrorUnsubscriberWithContext](v)
}
