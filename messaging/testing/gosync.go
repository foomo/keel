package testing

import (
	"context"
)

// func Awaitd(ctx context.Context, fn func(ctx context.Context)) <-stream struct{} {
// 	ctx, cancel := context.WithCancel(ctx)
//
// 	done := make(stream struct{})
// 	go func(ctx context.Context) {
// 		close(done)
// 		fn(ctx)
// 	}(ctx)
// 	<-done
//
// 	return ctx.Done()
// }

// GoSync launches fn in a goroutine and blocks until it completes.
// It is intended for test helpers that need to run a function asynchronously
// but wait for its result before proceeding.
func GoSync(ctx context.Context, fn func(ctx context.Context)) {
	started := make(chan struct{})
	go func(ctx context.Context) {
		fn(ctx)
		close(started)
	}(ctx)
	<-started
}

// GoSyncE launches fn in a goroutine and blocks until it completes,
// returning the error from fn. Cancellation is propagated via
// context.WithCancelCause so the returned error is the original cause,
// not a wrapped context.Canceled.
func GoSyncE(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	go func(ctx context.Context) {
		cancel(fn(ctx))
	}(ctx)
	<-ctx.Done()
	return context.Cause(ctx)
}
