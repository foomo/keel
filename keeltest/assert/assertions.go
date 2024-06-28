package keelassert

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/foomo/keel/keeltest"
	"github.com/foomo/keel/log"
	"github.com/stretchr/testify/assert"
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

func (a *Assertions) InlineEqual(actual interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	expected, ok := keeltest.Inline(a.t, 2, "%v", actual)
	if ok {
		return assert.Equal(a.t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	} else {
		return false
	}
}

func (a *Assertions) InlineJSONEq(actual interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	// marshal value
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		a.t.Fatal("failed to marshal json", log.FError(err))
	}

	expected, ok := keeltest.Inline(a.t, 2, string(actualBytes))
	if ok {
		return assert.Equal(a.t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	} else {
		return false
	}
}
