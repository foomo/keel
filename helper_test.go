package keel_test

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foomo/keel"
	"github.com/stretchr/testify/require"
)

// closerCall records a testCloser call and the state of its context.
type closerCall struct {
	name        string
	err         error
	hasDeadline bool
}

// testCloser implements interfaces.ErrorCloserWithContext and reports its calls.
type testCloser struct {
	name  string
	calls chan<- closerCall
	// block waits for the context to be done before reporting
	block bool
	// gate, if set, is waited for after reporting
	gate <-chan struct{}
	err  error
}

func (c *testCloser) Close(ctx context.Context) error {
	if c.block {
		<-ctx.Done()
	}

	_, hasDeadline := ctx.Deadline()
	c.calls <- closerCall{name: c.name, err: ctx.Err(), hasDeadline: hasDeadline}

	if c.gate != nil {
		<-c.gate
	}

	return c.err
}

type httpResult struct {
	statusCode int
	err        error
}

// recordCloser implements interfaces.ErrorCloserWithContext and records calls.
type recordCloser struct {
	closed atomic.Bool
}

func (c *recordCloser) Close(_ context.Context) error {
	c.closed.Store(true)
	return nil
}

func httpGet(ctx context.Context, url string) httpResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return httpResult{err: err}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return httpResult{err: err}
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	return httpResult{statusCode: resp.StatusCode}
}

// startServer runs the server and returns a channel closed once Run returned.
// The server is shut down on cleanup in case the test did not stop it.
func startServer(t *testing.T, svr *keel.Server) <-chan struct{} {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		svr.Run()
	}()

	t.Cleanup(func() {
		svr.ShutdownCancel()()
		receive(t, done, "server did not stop on cleanup")
	})

	return done
}

func requireEventuallyStatus(t *testing.T, url string, statusCode int) {
	t.Helper()

	require.Eventually(t, func() bool {
		res := httpGet(t.Context(), url)
		return res.err == nil && res.statusCode == statusCode
	}, 5*time.Second, 10*time.Millisecond, "%s did not respond with %d", url, statusCode)
}

// receive returns the next value of ch, failing t if none arrives in time.
func receive[T any](t *testing.T, ch <-chan T, msg string) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal(msg)
	}

	var zero T

	return zero
}
