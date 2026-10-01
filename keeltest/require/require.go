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

// InlineEqual compares the %v formatting of actual with the value stored
// as an INLINE comment on the calling line, see [keeltest.Inline]. A missing
// value is written and the test is marked as failed. It stops the test if
// the values differ.
func InlineEqual(t *testing.T, actual any, msgAndArgs ...any) {
	t.Helper()

	if expected, ok := keeltest.Inline(t, 2, "%v", actual); ok {
		require.Equal(t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	}
}

// InlineJSONEq compares the JSON encoding of actual with the JSON stored
// as an INLINE comment on the calling line, ignoring formatting, see
// [keeltest.Inline]. A missing value is written and the test is marked as
// failed. It stops the test if the values differ.
func InlineJSONEq(t *testing.T, actual any, msgAndArgs ...any) {
	t.Helper()
	// marshal value
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatal("failed to marshal json", log.FError(err))
	}

	if expected, ok := keeltest.Inline(t, 2, string(actualBytes)); ok {
		require.Equal(t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	}
}
