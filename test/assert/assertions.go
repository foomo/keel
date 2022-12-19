package keelassert

import (
	"testing"
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
	return InlineEqual(a.t, actual, msgAndArgs...)
}

func (a *Assertions) InlineJSONEq(actual interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return InlineJSONEq(a.t, actual, msgAndArgs...)
}
