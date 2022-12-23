package keelassert

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/foomo/keel/log"
	keeltestutil "github.com/foomo/keel/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/pretty"
)

func InlineEqual(t *testing.T, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	expected, ok := keeltestutil.Inline(t, actual, 2)
	if ok {
		return assert.Equal(t, expected, fmt.Sprintf("%v", actual), msgAndArgs...)
	} else {
		return false
	}
}

func InlineJSONEq(t *testing.T, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	// marshal value
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatal("failed to marshal json", log.FError(err))
	}

	expected, ok := keeltestutil.Inline(t, string(actualBytes), 2)
	if ok {
		return assert.Equal(t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	} else {
		return false
	}
}
