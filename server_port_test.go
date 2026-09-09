package keel_test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestServerPortInUse covers a pod restarting while its port is still held, e.g.
// after an OOMKill. The server must stop instead of staying alive with a service
// that never got a listener, so the container is restarted rather than reporting
// healthy while serving nothing.
func TestServerPortInUse(t *testing.T) {
	t.Parallel()

	l := zaptest.NewLogger(t)

	var lc net.ListenConfig

	occupied, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	require.NoError(t, err)

	t.Cleanup(func() { _ = occupied.Close() })

	svr := keel.NewServer(
		keel.WithContext(t.Context()),
		keel.WithLogger(l),
		keel.WithGracefulPeriod(3*time.Second),
	)

	blocked := service.NewHTTP(l, "blocked", occupied.Addr().String(), http.NewServeMux())
	svr.AddService(blocked)

	done := make(chan struct{})

	go func() {
		defer close(done)

		svr.Run()
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("server kept running although a service could not bind its port")
	}

	require.Error(t, blocked.Healthz(), "service must not report healthy without a listener")
	require.Error(t, svr.Healthz(), "server must not report healthy after it stopped")
}

// TestServerPortInUseStopsOtherServices asserts the whole server goes down, not
// just the service that failed, so a partially serving pod cannot pass its probes.
func TestServerPortInUseStopsOtherServices(t *testing.T) {
	t.Parallel()

	l := zaptest.NewLogger(t)

	var lc net.ListenConfig

	occupied, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	require.NoError(t, err)

	t.Cleanup(func() { _ = occupied.Close() })

	free, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	require.NoError(t, err)

	freeAddr := free.Addr().String()
	require.NoError(t, free.Close())

	svr := keel.NewServer(
		keel.WithContext(t.Context()),
		keel.WithLogger(l),
		keel.WithGracefulPeriod(3*time.Second),
	)

	healthy := service.NewHTTP(l, "healthy", freeAddr, http.NewServeMux())
	blocked := service.NewHTTP(l, "blocked", occupied.Addr().String(), http.NewServeMux())
	svr.AddServices(healthy, blocked)

	done := make(chan struct{})

	go func() {
		defer close(done)

		svr.Run()
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("server kept running although a service could not bind its port")
	}

	dialer := net.Dialer{Timeout: time.Second}

	require.Eventually(t, func() bool {
		conn, err := dialer.DialContext(t.Context(), "tcp", freeAddr)
		if err != nil {
			return true
		}

		_ = conn.Close()

		return false
	}, 5*time.Second, 50*time.Millisecond, "the remaining service must not keep serving")
}
