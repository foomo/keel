package telemetry_test

import (
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestMeter(t *testing.T) {
	t.Parallel()

	m := telemetry.Meter()
	assert.NotNil(t, m)
}
