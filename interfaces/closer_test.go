package interfaces_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

type closerPlain struct{ called *bool }

func (c closerPlain) Close() { *c.called = true }

type closerErr struct{}

func (closerErr) Close() error { return nil }

type closerCtx struct{}

func (closerCtx) Close(context.Context) {}

type closerErrCtx struct{}

func (closerErrCtx) Close(context.Context) error { return nil }

func TestIsCloser(t *testing.T) {
	t.Parallel()

	called := false
	if c, ok := interfaces.IsCloser(closerPlain{called: &called}); !ok {
		t.Fatal("expected satisfier")
	} else {
		c.Close()
	}

	if !called {
		t.Fatal("Close was not invoked")
	}

	if _, ok := interfaces.IsCloser(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorCloser(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorCloser(closerErr{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorCloser(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsCloserWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsCloserWithContext(closerCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsCloserWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorCloserWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorCloserWithContext(closerErrCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorCloserWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

// ExampleIsCloser shows the type-assertion dispatch pattern used to invoke
// whichever Close shape a service implements.
func ExampleIsCloser() {
	var svc any = closerErrCtx{}

	if _, ok := interfaces.IsCloser(svc); ok {
		fmt.Println("plain Closer")
	}

	if c, ok := interfaces.IsErrorCloserWithContext(svc); ok {
		fmt.Println(c.Close(context.Background()))
	}
	// Output: <nil>
}

// ExampleIsErrorCloserWithContext asserts for the richest Close shape and
// calls it with a bounded context.
func ExampleIsErrorCloserWithContext() {
	var v any = closerErrCtx{}

	c, ok := interfaces.IsErrorCloserWithContext(v)
	if !ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println(c.Close(ctx))
	// Output: <nil>
}
