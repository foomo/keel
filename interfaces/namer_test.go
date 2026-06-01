package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

type namer struct{ n string }

func (x namer) Name() string { return x.n }

func TestIsNamer(t *testing.T) {
	t.Parallel()

	got, ok := interfaces.IsNamer(namer{n: "queue"})
	if !ok {
		t.Fatal("expected satisfier")
	}

	if got.Name() != "queue" {
		t.Fatalf("Name()=%q, want %q", got.Name(), "queue")
	}

	if _, ok := interfaces.IsNamer(nothing{}); ok {
		t.Fatal("non-satisfier should miss")
	}
}

// ExampleIsNamer reports whether v carries a stable identifier.
func ExampleIsNamer() {
	var v any = namer{n: "queue"}

	if n, ok := interfaces.IsNamer(v); ok {
		fmt.Println(n.Name())
	}
	// Output: queue
}
