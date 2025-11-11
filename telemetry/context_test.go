package telemetry_test

import (
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/foomo/keel/telemetry/telemetrytest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestCtx(t *testing.T) {
	t.Parallel()

	var err error

	zap.ReplaceGlobals(zaptest.NewLogger(t,
		zaptest.WrapOptions(zap.AddCaller()),
		zaptest.Level(zap.DebugLevel),
	))

	spanRecorder, _ := telemetrytest.NewTestTraceProvider()

	_, err = telemetry.NewStdOutMeterProvider(t.Context())
	require.NoError(t, err)

	t.Run("Log", func(t *testing.T) {
		t.Parallel()
		ctx := telemetry.Ctx(t.Context())

		attr := attribute.String("foo", "bar")
		ctx.LogInfo("Info", attr)
		ctx.LogWarn("Warn", attr)
		ctx.LogDebug("Debug", attr)
		ctx.LogError("Error", attr)
	})

	t.Run("StartSpan", func(t *testing.T) { //nolint:paralleltest
		spanRecorder.Reset()

		ctx := telemetry.Ctx(t.Context()).StartSpan()
		span := ctx.Span()

		require.Len(t, spanRecorder.Started(), 1)
		span.End()
		require.Len(t, spanRecorder.Ended(), 1)
	})

	t.Run("EndSpan Ok", func(t *testing.T) { //nolint:paralleltest
		spanRecorder.Reset()

		ctx := telemetry.Ctx(t.Context()).StartSpan()
		require.Len(t, spanRecorder.Started(), 1)
		ctx.EndSpan(nil)
		// repeat
		ctx.EndSpan(errors.New("error"))

		spans := spanRecorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Ok, spans[0].Status().Code)
	})

	t.Run("EndSpan Error", func(t *testing.T) { //nolint:paralleltest
		spanRecorder.Reset()

		ctx := telemetry.Ctx(t.Context()).StartSpan()
		require.Len(t, spanRecorder.Started(), 1)
		ctx.EndSpan(errors.New("error"))
		// repeat
		ctx.EndSpan(nil)

		spans := spanRecorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status().Code)
	})
}
