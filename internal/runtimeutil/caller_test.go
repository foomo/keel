package runtimeutil_test

import (
	"strings"
	"testing"

	"github.com/foomo/keel/internal/runtimeutil"
	"github.com/stretchr/testify/assert"
)

type caller struct{}

func (c *caller) caller() (string, string, string, int, bool) {
	return runtimeutil.Caller(0)
}

func TestCaller(t *testing.T) {
	t.Parallel()

	shortName, fullName, file, line, _ := runtimeutil.Caller(0)

	assert.Equal(t, "runtimeutil_test.TestCaller", shortName)
	assert.Equal(t, "github.com/foomo/keel/internal/runtimeutil_test.TestCaller", fullName)
	assert.True(t, strings.HasSuffix(file, "github.com/foomo/keel/internal/runtimeutil/caller_test.go"))
	assert.Equal(t, 20, line)

	c := new(caller)
	shortName, fullName, file, line, _ = c.caller()

	assert.Equal(t, "runtimeutil_test.(*caller).caller", shortName)
	assert.Equal(t, "github.com/foomo/keel/internal/runtimeutil_test.(*caller).caller", fullName)
	assert.True(t, strings.HasSuffix(file, "github.com/foomo/keel/internal/runtimeutil/caller_test.go"))
	assert.Equal(t, 14, line)
}
