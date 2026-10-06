package nats_test

import (
	"context"
	"testing"
	"time"

	testingx "github.com/foomo/go/testing"
	keelnats "github.com/foomo/keel/integration/nats"
	"github.com/foomo/keel/integration/nats/service"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type testRuntime struct {
	ctx   context.Context
	l     *zap.Logger
	meter metric.Meter
}

func (r *testRuntime) Logger() *zap.Logger      { return r.l }
func (r *testRuntime) Config() *viper.Viper     { return viper.New() }
func (r *testRuntime) Context() context.Context { return r.ctx }
func (r *testRuntime) Meter() metric.Meter      { return r.meter }
func (r *testRuntime) Tracer() trace.Tracer     { return noop.NewTracerProvider().Tracer("") }
func (r *testRuntime) AddCloser(any)            {}
func (r *testRuntime) AddClosers(...any)        {}

// TestConnect_disconnectServerAttrs asserts that the disconnect metric and log
// carry the server address although the connection is no longer connected.
func TestConnect_disconnectServerAttrs(t *testing.T) {
	t.Parallel()

	srv, err := service.NewEmbeddedServer(
		service.EmbeddedServerWithPort(testingx.FreePort(t)),
		service.EmbeddedServerWithHost("localhost"),
	)
	require.NoError(t, err)

	go func() { _ = srv.Start(t.Context()) }()

	require.Eventually(t, srv.Server().Running, time.Second, 10*time.Millisecond)

	reader := sdkmetric.NewManualReader()
	core, logs := observer.New(zap.DebugLevel)
	rt := &testRuntime{
		ctx:   t.Context(),
		l:     zap.New(core),
		meter: sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test"),
	}

	conn, err := keelnats.Connect(rt, srv.ClientURL())
	require.NoError(t, err)
	t.Cleanup(conn.Close)

	require.NoError(t, srv.Close(t.Context()))

	require.Eventually(t, func() bool {
		return logs.FilterMessage("disconnected").Len() > 0
	}, 5*time.Second, 10*time.Millisecond)

	assert.Equal(t, "localhost", logs.FilterMessage("disconnected").All()[0].ContextMap()["server_address"])

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))

	var addr attribute.Value

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if sum, ok := m.Data.(metricdata.Sum[int64]); ok && m.Name == "nats.client.disconnects" {
				addr, _ = sum.DataPoints[0].Attributes.Value(semconv.ServerAddressKey)
			}
		}
	}

	assert.Equal(t, "localhost", addr.AsString())
}
