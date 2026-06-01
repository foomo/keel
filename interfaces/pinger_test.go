package interfaces_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

type pingerErr struct{ err error }

func (p pingerErr) Ping() error { return p.err }

type pingerErrCtx struct{ err error }

func (p pingerErrCtx) Ping(context.Context) error { return p.err }

func TestIsErrorPinger(t *testing.T) {
	t.Parallel()

	want := errors.New("unreachable")

	p, ok := interfaces.IsErrorPinger(pingerErr{err: want})
	if !ok {
		t.Fatal("expected satisfier")
	}

	if got := p.Ping(); !errors.Is(got, want) {
		t.Fatalf("Ping()=%v, want %v", got, want)
	}

	if _, ok := interfaces.IsErrorPinger(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

func TestIsErrorPingerWithContext(t *testing.T) {
	t.Parallel()

	if _, ok := interfaces.IsErrorPingerWithContext(pingerErrCtx{}); !ok {
		t.Fatal("expected satisfier")
	}

	if _, ok := interfaces.IsErrorPingerWithContext(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

// ExampleIsErrorPinger asserts the health-probe shape.
func ExampleIsErrorPinger() {
	var v any = pingerErr{err: errors.New("unreachable")}

	if p, ok := interfaces.IsErrorPinger(v); ok {
		fmt.Println(p.Ping())
	}
	// Output: unreachable
}
