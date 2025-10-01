package telemetry_test

import (
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResource(t *testing.T) {
	result, err := telemetry.NewResource(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, result)
}
