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

func InlineEqual(t *testing.T, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	expected, ok := keeltest.Inline(t, 2, "%v", actual)
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

	expected, ok := keeltest.Inline(t, 2, string(actualBytes))
	if ok {
		return assert.Equal(t, string(pretty.Pretty([]byte(expected))), string(pretty.Pretty(actualBytes)), msgAndArgs...)
	} else {
		return false
	}
}
