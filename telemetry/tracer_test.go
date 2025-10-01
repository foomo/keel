package telemetry_test

import (
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestTracer(t *testing.T) {
	m := telemetry.Tracer()
	assert.NotNil(t, m)
}
