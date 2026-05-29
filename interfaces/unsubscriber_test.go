package interfaces_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

type unsubscriberPlain struct{}

func (unsubscriberPlain) Unsubscribe() {}

type unsubscriberErr struct{}

func (unsubscriberErr) Unsubscribe() error { return nil }

type unsubscriberCtx struct{ called *bool }

func (u unsubscriberCtx) Unsubscribe(context.Context) { *u.called = true }

type unsubscriberErrCtx struct{}

func (unsubscriberErrCtx) Unsubscribe(context.Context) error { return nil }

func TestIsUnsubscriber(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsUnsubscriber(unsubscriberPlain{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsUnsubscriber(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorUnsubscriber(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorUnsubscriber(unsubscriberErr{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorUnsubscriber(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsUnsubscriberWithContext(t *testing.T) {
	t.Parallel()

	called := false
	if u, ok := interfaces.IsUnsubscriberWithContext(unsubscriberCtx{called: &called}); !ok {
		t.Fatal("expected satisfier")
	} else {
		u.Unsubscribe(context.Background())
	}

	if !called {
		t.Fatal("Unsubscribe was not invoked")
	}

	if _, ok := interfaces.IsUnsubscriberWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorUnsubscriberWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorUnsubscriberWithContext(unsubscriberErrCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorUnsubscriberWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

// ExampleIsUnsubscriberWithContext asserts the context-aware Unsubscribe
// shape.
func ExampleIsUnsubscriberWithContext() {
	called := false

	var v any = unsubscriberCtx{called: &called}

	if u, ok := interfaces.IsUnsubscriberWithContext(v); ok {
		u.Unsubscribe(context.Background())
		fmt.Println(called)
	}
	// Output: true
}
