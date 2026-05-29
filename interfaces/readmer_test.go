package interfaces_test

import (
	"fmt"
	"testing"

	"github.com/foomo/keel/interfaces"
)

func TestReadmeFunc(t *testing.T) {
	t.Parallel()

	r := interfaces.ReadmeFunc(func() string { return "service: cache" })
	if got := r.Readme(); got != "service: cache" {
		t.Fatalf("Readme()=%q", got)
	}

	// ReadmeHandler must satisfy Readmer.
	var _ interfaces.Readmer = r
}

// ExampleReadmeFunc adapts a func returning documentation into a
// [interfaces.Readmer].
func ExampleReadmeFunc() {
	r := interfaces.ReadmeFunc(func() string { return "service: cache" })

	fmt.Println(r.Readme())
	// Output: service: cache
}
