package interfaces_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

type shutdownerPlain struct{ called *bool }

func (s shutdownerPlain) Shutdown() { *s.called = true }

type shutdownerErr struct{}

func (shutdownerErr) Shutdown() error { return nil }

type shutdownerCtx struct{}

func (shutdownerCtx) Shutdown(context.Context) {}

type shutdownerErrCtx struct{}

func (shutdownerErrCtx) Shutdown(context.Context) error { return nil }

func TestIsShutdowner(t *testing.T) {
	t.Parallel()

	called := false
	if s, ok := interfaces.IsShutdowner(shutdownerPlain{called: &called}); !ok {
		t.Fatal("expected satisfier")
	} else {
		s.Shutdown()
	}

	if !called {
		t.Fatal("Shutdown was not invoked")
	}

	if _, ok := interfaces.IsShutdowner(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorShutdowner(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorShutdowner(shutdownerErr{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorShutdowner(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsShutdownerWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsShutdownerWithContext(shutdownerCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsShutdownerWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorShutdownerWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorShutdownerWithContext(shutdownerErrCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorShutdownerWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

// ExampleIsShutdowner asserts the simplest Shutdown shape.
func ExampleIsShutdowner() {
	called := false

	var v any = shutdownerPlain{called: &called}

	if s, ok := interfaces.IsShutdowner(v); ok {
		s.Shutdown()
		fmt.Println(called)
	}
	// Output: true
}
