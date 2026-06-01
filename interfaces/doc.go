// Package interfaces defines small structural interface contracts used
// throughout keel to describe the lifecycle and identity of services and
// resources: closing, shutting down, stopping, unsubscribing, pinging,
// naming, and self-documenting.
//
// Each lifecycle family (Closer, Shutdowner, Stopper, Unsubscriber, Pinger)
// is declared along four shape axes covering the common Go patterns:
//
//   - the unadorned method (no return, no context),
//   - the variant that returns an error,
//   - the variant that accepts a [context.Context],
//   - the variant that accepts a [context.Context] and returns an error.
//
// For every named interface X declared here there is a paired helper
// IsX(v any) (X, bool) that performs the type assertion. All such helpers
// are thin wrappers around the generic [Is] and exist so callers can keep
// a uniform dispatch style — see the [github.com/foomo/keel] server's
// service shutdown loop for the canonical usage.
package interfaces

// Is reports whether v implements T and returns the asserted value.
// When v does not implement T the returned value is the zero value of T
// and ok is false.
func Is[T any](v any) (T, bool) {
	t, ok := v.(T)
	return t, ok
}
