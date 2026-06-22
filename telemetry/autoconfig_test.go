package telemetry_test

import (
	"context"
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTLPProtocol(t *testing.T) {
	t.Run("defaults to http/protobuf", func(t *testing.T) {
		assert.Equal(t, "http/protobuf", telemetry.OTLPProtocol("traces"))
	})

	t.Run("global override", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
		assert.Equal(t, "grpc", telemetry.OTLPProtocol("traces"))
	})

	t.Run("per-signal override wins over global", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
		t.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http/protobuf")
		assert.Equal(t, "http/protobuf", telemetry.OTLPProtocol("metrics"))
		assert.Equal(t, "grpc", telemetry.OTLPProtocol("traces"))
	})
}

func TestNewTraceProviderFromEnv(t *testing.T) {
	t.Run("none by default", func(t *testing.T) {
		tp, err := telemetry.NewTraceProviderFromEnv(context.Background())
		require.NoError(t, err)
		assert.Nil(t, tp)
	})

	t.Run("console", func(t *testing.T) {
		t.Setenv("OTEL_TRACES_EXPORTER", "console")

		tp, err := telemetry.NewTraceProviderFromEnv(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, tp)
	})

	t.Run("unknown exporter errors", func(t *testing.T) {
		t.Setenv("OTEL_TRACES_EXPORTER", "bogus")

		tp, err := telemetry.NewTraceProviderFromEnv(context.Background())
		require.Error(t, err)
		assert.Nil(t, tp)
	})

	t.Run("otlp unknown protocol errors", func(t *testing.T) {
		t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
		t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "bogus")

		tp, err := telemetry.NewTraceProviderFromEnv(context.Background())
		require.Error(t, err)
		assert.Nil(t, tp)
	})
}

func TestNewMeterProviderFromEnv(t *testing.T) {
	t.Run("none by default", func(t *testing.T) {
		mp, err := telemetry.NewMeterProviderFromEnv(context.Background())
		require.NoError(t, err)
		assert.Nil(t, mp)
	})

	t.Run("prometheus", func(t *testing.T) {
		t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")

		mp, err := telemetry.NewMeterProviderFromEnv(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, mp)
	})

	t.Run("unknown exporter errors", func(t *testing.T) {
		t.Setenv("OTEL_METRICS_EXPORTER", "bogus")

		mp, err := telemetry.NewMeterProviderFromEnv(context.Background())
		require.Error(t, err)
		assert.Nil(t, mp)
	})

	t.Run("otlp unknown protocol errors", func(t *testing.T) {
		t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
		t.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "bogus")

		mp, err := telemetry.NewMeterProviderFromEnv(context.Background())
		require.Error(t, err)
		assert.Nil(t, mp)
	})
}

func TestNewLoggerProviderFromEnv(t *testing.T) {
	t.Run("none by default", func(t *testing.T) {
		lp, err := telemetry.NewLoggerProviderFromEnv(context.Background())
		require.NoError(t, err)
		assert.Nil(t, lp)
	})

	t.Run("console", func(t *testing.T) {
		t.Setenv("OTEL_LOGS_EXPORTER", "console")

		lp, err := telemetry.NewLoggerProviderFromEnv(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, lp)
	})

	t.Run("unknown exporter errors", func(t *testing.T) {
		t.Setenv("OTEL_LOGS_EXPORTER", "bogus")

		lp, err := telemetry.NewLoggerProviderFromEnv(context.Background())
		require.Error(t, err)
		assert.Nil(t, lp)
	})

	t.Run("otlp unknown protocol errors", func(t *testing.T) {
		t.Setenv("OTEL_LOGS_EXPORTER", "otlp")
		t.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "bogus")

		lp, err := telemetry.NewLoggerProviderFromEnv(context.Background())
		require.Error(t, err)
		assert.Nil(t, lp)
	})
}
