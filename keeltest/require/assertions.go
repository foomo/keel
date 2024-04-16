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

// Assertions provides assertion methods around the
// TestingT interface.
type Assertions struct {
	t *testing.T
}

// New makes a new Assertions object for the specified TestingT.
func New(t *testing.T) *Assertions { //nolint:thelper
	return &Assertions{
		t: t,
	}
}

func (a *Assertions) InlineEqual(actual interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if expected, ok := keeltest.Inline(a.t, 2, "%v", actual); ok {
		require.Equal(a.t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	}
}

func (a *Assertions) InlineJSONEq(actual interface{}, msgAndArgs ...interface{}) {
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
