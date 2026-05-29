package interfaces_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

func TestCloserFunc(t *testing.T) {
	t.Parallel()

	want := errors.New("boom")
	got := interfaces.CloserFunc(func(context.Context) error { return want }).Close(context.Background())

	if !errors.Is(got, want) {
		t.Fatalf("Close returned %v, want %v", got, want)
	}

	// CloserFunc must satisfy the richest Close shape.
	if _, ok := interfaces.IsErrorCloserWithContext(interfaces.CloserFunc(func(context.Context) error { return nil })); !ok {
		t.Fatal("CloserFunc should satisfy ErrorCloserWithContext")
	}
}

// ExampleCloserFunc adapts a plain func into an
// [interfaces.ErrorCloserWithContext].
func ExampleCloserFunc() {
	closer := interfaces.CloserFunc(func(context.Context) error {
		fmt.Println("closing")
		return nil
	})

	_ = closer.Close(context.Background())
	// Output: closing
}
