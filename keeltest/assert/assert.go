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

// InlineEqual compares the %v formatting of actual with the value stored
// as an INLINE comment on the calling line, see [keeltest.Inline]. A missing
// value is written and the test is marked as failed. It reports whether
// the values are equal.
func InlineEqual(t *testing.T, actual any, msgAndArgs ...any) bool {
	t.Helper()

	expected, ok := keeltest.Inline(t, 2, "%v", actual)
	if ok {
		return assert.Equal(t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	} else {
		return false
	}
}

// InlineJSONEq compares the JSON encoding of actual with the JSON stored
// as an INLINE comment on the calling line, ignoring formatting, see
// [keeltest.Inline]. A missing value is written and the test is marked as
// failed. It reports whether the values are equal.
func InlineJSONEq(t *testing.T, actual any, msgAndArgs ...any) bool {
	t.Helper()
	// marshal value
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatal("failed to marshal json", log.FError(err))
	}

	expected, ok := keeltest.Inline(t, 2, string(actualBytes))
	if ok {
		return assert.Equal(t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	} else {
		return false
	}
}
