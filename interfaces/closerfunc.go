package interfaces

import (
	"context"
)

// CloserFunc is an adapter that allows an ordinary function with the
// signature func(ctx context.Context) error to be used as an
// [ErrorCloserWithContext]. It mirrors the pattern of [net/http.HandlerFunc].
type CloserFunc func(ctx context.Context) error

// Close calls f(ctx) and returns its result, so a CloserFunc satisfies
// [ErrorCloserWithContext].
func (f CloserFunc) Close(ctx context.Context) error {
	return f(ctx)
}
