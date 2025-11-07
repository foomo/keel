package telemetry_test

import (
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNewNoopTraceProvider(t *testing.T) {
	t.Parallel()

	tp := telemetry.NewNoopTraceProvider()
	assert.IsType(t, noop.TracerProvider{}, tp)
}
