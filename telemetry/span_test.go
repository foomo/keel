package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/foomo/keel/telemetry"
	"github.com/foomo/keel/telemetry/telemetrytest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// withEndSpan is the WRONG way: `err` is evaluated when the defer statement is
// reached, which is before the return assigns the error. The span never learns
// about it.
func withEndSpan(ctx context.Context) (err error) { //nolint:nonamedreturns
	_, sp := telemetry.StartSpan(ctx)
	defer telemetry.EndSpan(sp, err)

	return errors.New("boom")
}

// withDeferEndSpan is the RIGHT way: the pointer is dereferenced when the
// deferred call actually runs, so it sees the final value of the named return.
func withDeferEndSpan(ctx context.Context) (err error) { //nolint:nonamedreturns
	_, sp := telemetry.StartSpan(ctx)
	defer telemetry.DeferEndSpan(sp, &err)

	return errors.New("boom")
}

// TestEndSpanVsDeferEndSpan showcases why DeferEndSpan exists: EndSpan takes the
// error by value, DeferEndSpan takes it by pointer.
func TestEndSpanVsDeferEndSpan(t *testing.T) {
	// NOTE: no t.Parallel(), NewTestTraceProvider mutates the global otel provider.
	spanRecorder, _ := telemetrytest.NewTestTraceProvider()

	endedSpan := func(t *testing.T) sdktrace.ReadOnlySpan {
		t.Helper()

		spans := spanRecorder.Ended()
		require.Len(t, spans, 1)

		return spans[0]
	}

	t.Run("EndSpan deferred loses the error", func(t *testing.T) {
		spanRecorder.Reset()

		require.Error(t, withEndSpan(t.Context()))

		span := endedSpan(t)
		assert.Equal(t, codes.Unset, span.Status().Code)
		assert.Empty(t, span.Events())
	})

	t.Run("DeferEndSpan records the error", func(t *testing.T) {
		spanRecorder.Reset()

		require.Error(t, withDeferEndSpan(t.Context()))

		span := endedSpan(t)
		assert.Equal(t, codes.Error, span.Status().Code)
		assert.Equal(t, "boom", span.Status().Description)
		require.Len(t, span.Events(), 1)
		assert.Equal(t, "exception", span.Events()[0].Name)
		assert.Equal(t, "telemetry_test.withDeferEndSpan", span.Name())
		assert.Contains(t, span.Attributes(), semconv.ErrorType(errors.New("boom")))
	})

	t.Run("EndSpan called directly records the error", func(t *testing.T) {
		spanRecorder.Reset()

		_, sp := telemetry.StartSpan(t.Context())
		telemetry.EndSpan(sp, errors.New("boom"))

		span := endedSpan(t)
		assert.Equal(t, codes.Error, span.Status().Code)
		assert.Equal(t, "boom", span.Status().Description)
	})

	t.Run("EndSpan with nil error leaves the status unset", func(t *testing.T) {
		spanRecorder.Reset()

		_, sp := telemetry.StartSpan(t.Context())
		telemetry.EndSpan(sp, nil)

		// Unlike the deprecated End, EndSpan does not set codes.Ok.
		assert.Equal(t, codes.Unset, endedSpan(t).Status().Code)
	})

	t.Run("DeferEndSpan with a nil pointer does not panic", func(t *testing.T) {
		spanRecorder.Reset()

		_, sp := telemetry.StartSpan(t.Context())
		require.NotPanics(t, func() { telemetry.DeferEndSpan(sp, nil) })

		assert.Equal(t, codes.Unset, endedSpan(t).Status().Code)
	})

	t.Run("DeferEndSpan forwards the end options", func(t *testing.T) {
		spanRecorder.Reset()

		endTime := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)

		var err error

		_, sp := telemetry.StartSpan(t.Context())
		telemetry.DeferEndSpan(sp, &err, trace.WithTimestamp(endTime))

		assert.Equal(t, endTime, endedSpan(t).EndTime())
	})
}
