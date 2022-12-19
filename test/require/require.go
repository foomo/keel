package keelrequire

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/foomo/keel/log"
	keeltestutil "github.com/foomo/keel/test/util"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/pretty"
)

func InlineEqual(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if t.Failed() {
		return
	}
	if expected, ok := keeltestutil.Inline(t, actual, 2); ok {
		require.Equal(t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	}
}

func InlineJSONEq(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if t.Failed() {
		return
	}
	// marshal value
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatal("failed to marshal json", log.FError(err))
	}
	if expected, ok := keeltestutil.Inline(t, string(actualBytes), 2); ok {
		require.Equal(t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	}
}
