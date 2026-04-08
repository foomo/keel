package testing

import (
	"context"
)

// GoAsync launches fn in a goroutine, waits for it to start, and returns
// a channel that is closed when fn calls done. Use this when the goroutine
// should keep running after launch and the caller needs to wait for an
// explicit completion signal later.
func GoAsync(ctx context.Context, fn func(ctx context.Context, done context.CancelFunc)) <-chan struct{} {
	ctx, cancel := context.WithCancel(ctx)

	started := make(chan struct{})
	go func(ctx context.Context, cancel context.CancelFunc) {
		close(started)
		fn(ctx, cancel)
	}(ctx, cancel)
	<-started

	return ctx.Done()
}

// GoAsyncE launches fn in a goroutine, waits for it to start, and returns
// a channel that receives the error from fn when it completes. Use this when
// the goroutine should keep running after launch and the caller needs to
// collect both the completion signal and the error later.
func GoAsyncE(ctx context.Context, fn func(ctx context.Context) error) <-chan error {
	errCh := make(chan error, 1)
	started := make(chan struct{})
	go func(ctx context.Context) {
		close(started)
		errCh <- fn(ctx)
	}(ctx)
	<-started

	return errCh
}
