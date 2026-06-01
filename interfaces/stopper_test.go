package interfaces_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

type stopperPlain struct{}

func (stopperPlain) Stop() {}

type stopperErr struct{}

func (stopperErr) Stop() error { return nil }

type stopperCtx struct{}

func (stopperCtx) Stop(context.Context) {}

type stopperErrCtx struct{}

func (stopperErrCtx) Stop(context.Context) error { return nil }

func TestIsStopper(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsStopper(stopperPlain{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsStopper(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorStopper(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorStopper(stopperErr{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorStopper(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsStopperWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsStopperWithContext(stopperCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsStopperWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorStopperWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorStopperWithContext(stopperErrCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorStopperWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

// ExampleIsErrorStopper asserts the error-returning Stop shape.
func ExampleIsErrorStopper() {
	var v any = stopperErr{}

	if s, ok := interfaces.IsErrorStopper(v); ok {
		fmt.Println(s.Stop())
	}
	// Output: <nil>
}
