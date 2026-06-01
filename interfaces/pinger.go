package interfaces

import "context"

// ErrorPinger is implemented by any value whose health can be probed
// through a Ping method that returns an error on failure. See also
// [ErrorPingerWithContext].
type ErrorPinger interface {
	Ping() error
}

// IsErrorPinger reports whether v implements [ErrorPinger] and returns the
// asserted value. The boolean is false when v does not implement
// [ErrorPinger].
func IsErrorPinger(v any) (ErrorPinger, bool) {
	return Is[ErrorPinger](v)
}

// ErrorPingerWithContext is like [ErrorPinger] but its Ping method accepts
// a [context.Context] so the caller can bound the probe.
type ErrorPingerWithContext interface {
	Ping(ctx context.Context) error
}

// IsErrorPingerWithContext reports whether v implements
// [ErrorPingerWithContext] and returns the asserted value. The boolean is
// false when v does not implement [ErrorPingerWithContext].
func IsErrorPingerWithContext(v any) (ErrorPingerWithContext, bool) {
	return Is[ErrorPingerWithContext](v)
}
