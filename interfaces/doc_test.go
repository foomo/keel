package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

// nothing implements none of the interfaces in this package and is shared
// across the per-family test files as the canonical non-satisfier.
type nothing struct{}

// ExampleIs shows the generic helper used directly to assert an arbitrary
// interface — here [interfaces.Readmer].
func ExampleIs() {
	var v any = interfaces.ReadmeFunc(func() string { return "ok" })

	if r, ok := interfaces.Is[interfaces.Readmer](v); ok {
		fmt.Println(r.Readme())
	}
	// Output: ok
}

// TestIsZeroValueOnMiss locks the generic helper's contract: when v does
// not implement T the returned value is the zero value of T.
func TestIsZeroValueOnMiss(t *testing.T) {
	t.Parallel()

	got, ok := interfaces.Is[interfaces.Closer](nothing{})
	if ok {
		t.Fatalf("expected ok=false")
	}

	if got != nil {
		t.Fatalf("expected zero Closer (nil), got %v", got)
	}
}
