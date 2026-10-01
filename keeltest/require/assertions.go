package keelrequire

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/foomo/keel/keeltest"
	"github.com/foomo/keel/log"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/pretty"
)

// Assertions provides the package's inline assertions bound to a
// [testing.T].
type Assertions struct {
	t *testing.T
}

// New returns [Assertions] for t.
func New(t *testing.T) *Assertions { //nolint:thelper
	return &Assertions{
		t: t,
	}
}

// InlineEqual is like the package level [InlineEqual].
func (a *Assertions) InlineEqual(actual any, msgAndArgs ...any) {
	a.t.Helper()

	if expected, ok := keeltest.Inline(a.t, 2, "%v", actual); ok {
		require.Equal(a.t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	}
}

// InlineJSONEq is like the package level [InlineJSONEq].
func (a *Assertions) InlineJSONEq(actual any, msgAndArgs ...any) {
	a.t.Helper()
	// marshal value
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		a.t.Fatal("failed to marshal json", log.FError(err))
	}

	if expected, ok := keeltest.Inline(a.t, 2, string(actualBytes)); ok {
		require.Equal(a.t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	}
}
