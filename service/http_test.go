package service_test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestHTTPStartPortInUse covers the case of a restarted pod whose predecessor
// still holds the port: Start must fail and the service must never report
// healthy, or the pod would come up without ever accepting a request.
func TestHTTPStartPortInUse(t *testing.T) {
	t.Parallel()

	l := zaptest.NewLogger(t)

	var lc net.ListenConfig

	occupied, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	require.NoError(t, err)

	t.Cleanup(func() { _ = occupied.Close() })

	s := service.NewHTTP(l, "blocked", occupied.Addr().String(), http.NewServeMux())

	require.Error(t, s.Healthz(), "service must not be healthy before it started")
	require.Error(t, s.Start(t.Context()), "start must fail while the port is occupied")
	require.ErrorIs(t, s.Healthz(), service.ErrServiceNotRunning,
		"service must not report healthy when it never got a listener")
}

// TestHTTPStartFreePort guards the fix against regressing the happy path.
func TestHTTPStartFreePort(t *testing.T) {
	t.Parallel()

	l := zaptest.NewLogger(t)

	var lc net.ListenConfig

	free, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	require.NoError(t, err)

	addr := free.Addr().String()
	require.NoError(t, free.Close())

	s := service.NewHTTP(l, "free", addr, http.NewServeMux())

	done := make(chan error, 1)
	go func() { done <- s.Start(t.Context()) }()

	require.Eventually(t, func() bool { return s.Healthz() == nil }, 5*time.Second, 10*time.Millisecond,
		"service must report healthy once it is listening")

	require.NoError(t, s.Close(t.Context()))
	require.NoError(t, <-done)
}

func _ExampleNewHTTP() {
	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
		keel.WithGracefulPeriod(10*time.Second),
	)

	l := svr.Logger()

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})),
	)

	go func() {
		waitFor("localhost:8080")
		l.Info(httpGet("http://localhost:8080"))
		shutdown()
	}()

	svr.Run()

	// Output:
	// {"level":"info","msg":"starting keel server"}
	// {"level":"info","msg":"starting keel service","keel_service_type":"http","keel_service_name":"demo","net_host_ip":"localhost","net_host_port":"8080"}
	// {"level":"info","msg":"OK"}
	// {"level":"info","msg":"keel closer closed","graceful_period":"10s"}
	// {"level":"info","msg":"keel closer closed: closers"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"http","keel_service_name":"demo"}
	// {"level":"debug","msg":"keel closer closed","name":"*service.HTTP"}
	// {"level":"debug","msg":"keel closer closed","name":"noop.TracerProvider"}
	// {"level":"debug","msg":"keel closer closed","name":"noop.MeterProvider"}
	// {"level":"debug","msg":"keel closer closed","name":"noop.LoggerProvider"}
	// {"level":"info","msg":"keel closer closed: complete"}
	// {"level":"info","msg":"keel server stopped"}
}
