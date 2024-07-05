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

func InlineEqual(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if expected, ok := keeltest.Inline(t, 2, "%v", actual); ok {
		require.Equal(t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	}
}

func InlineJSONEq(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
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
