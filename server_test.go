package keel_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	testingx "github.com/foomo/go/testing"
	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// shutdownSignal is only handled by the suite servers, so sending it does not
// stop servers of tests running in parallel.
const shutdownSignal = syscall.SIGUSR2

type KeelTestSuite struct {
	suite.Suite
	l        *zap.Logger
	svr      *keel.Server
	mux      *http.ServeMux
	addr     string
	zapAddr  string
	sleeping chan struct{}
	done     chan struct{}
	cancel   context.CancelFunc
}

// BeforeTest hook
func (s *KeelTestSuite) BeforeTest(suiteName, testName string) {
	ports := testingx.FreePorts(s.T(), 2)
	s.addr = fmt.Sprintf("localhost:%d", ports[0])
	s.zapAddr = fmt.Sprintf("localhost:%d", ports[1])
	s.sleeping = make(chan struct{})
	s.done = nil

	s.l = zaptest.NewLogger(s.T())
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s.mux.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		close(s.sleeping)
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	s.mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("foobar")
	})
	s.mux.HandleFunc("/log/info", func(w http.ResponseWriter, r *http.Request) {
		s.l.Info("logging info")
	})
	s.mux.HandleFunc("/log/debug", func(w http.ResponseWriter, r *http.Request) {
		s.l.Debug("logging debug")
	})
	s.mux.HandleFunc("/log/warn", func(w http.ResponseWriter, r *http.Request) {
		s.l.Warn("logging warn")
	})
	s.mux.HandleFunc("/log/error", func(w http.ResponseWriter, r *http.Request) {
		s.l.Error("logging error")
	})

	ctx, cancel := context.WithCancel(s.T().Context())
	s.svr = keel.NewServer(
		keel.WithContext(ctx),
		keel.WithLogger(s.l),
		keel.WithShutdownSignals(shutdownSignal),
	)
	s.cancel = cancel
}

// AfterTest hook
func (s *KeelTestSuite) AfterTest(suiteName, testName string) {
	s.cancel()
	s.waitServerStopped()
}

func (s *KeelTestSuite) TestServiceHTTP() {
	s.svr.AddServices(
		service.NewHTTP(s.l, "test", s.addr, s.mux),
	)

	s.runServer()

	if statusCode, _, err := s.httpGet(s.url("/ok")); s.NoError(err) {
		s.Equal(http.StatusOK, statusCode)
	}
}

func (s *KeelTestSuite) TestServiceHTTPZap() {
	s.svr.AddServices(
		service.NewHTTPZap(s.l, "zap", s.zapAddr, "/log"),
		service.NewHTTP(s.l, "test", s.addr, s.mux),
	)

	s.runServer()

	zapURL := "http://" + s.zapAddr + "/log"

	// the zap settings are global, restore them for following runs
	defer func() {
		_, _, err := s.httpPut(zapURL, `{"level":"info","disableCaller":true,"disableStacktrace":true}`)
		s.NoError(err)
	}()

	s.Run("default", func() {
		if statusCode, body, err := s.httpGet(zapURL); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"info","disableCaller":true,"disableStacktrace":true}`, body)
		}

		if statusCode, _, err := s.httpGet(s.url("/log/info")); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}

		if statusCode, _, err := s.httpGet(s.url("/log/debug")); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("set debug level", func() {
		if statusCode, body, err := s.httpPut(zapURL, `{"level":"debug"}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"debug","disableCaller":true,"disableStacktrace":true}`, body)
		}

		if statusCode, _, err := s.httpGet(s.url("/log/info")); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}

		if statusCode, _, err := s.httpGet(s.url("/log/debug")); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("enable caller", func() {
		if statusCode, body, err := s.httpPut(zapURL, `{"disableCaller":false}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"debug","disableCaller":false,"disableStacktrace":true}`, body)
		}

		if statusCode, _, err := s.httpGet(s.url("/log/error")); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("enable stacktrace", func() {
		if statusCode, body, err := s.httpPut(zapURL, `{"disableStacktrace":false}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"debug","disableCaller":false,"disableStacktrace":false}`, body)
		}

		if statusCode, _, err := s.httpGet(s.url("/log/error")); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})
}

func (s *KeelTestSuite) TestGraceful() {
	s.svr.AddServices(
		service.NewHTTP(s.l, "test", s.addr, s.mux),
	)

	s.runServer()

	type result struct {
		statusCode int
		err        error
	}

	// start long running request
	sleepResult := make(chan result, 1)

	go func() {
		statusCode, _, err := s.httpGet(s.url("/sleep"))
		sleepResult <- result{statusCode: statusCode, err: err}
	}()

	select {
	case <-s.sleeping:
	case <-time.After(5 * time.Second):
		s.FailNow("request to /sleep did not arrive")
	}

	// shutdown while the request is in flight
	s.Require().NoError(syscall.Kill(syscall.Getpid(), shutdownSignal))
	s.waitServerStopped()

	// in flight request must complete
	select {
	case res := <-sleepResult:
		if s.NoError(res.err) {
			s.Equal(http.StatusOK, res.statusCode)
		}
	case <-time.After(5 * time.Second):
		s.FailNow("request to /sleep did not complete")
	}

	// server must be down
	_, _, err := s.httpGet(s.url("/ok"))
	s.Require().Error(err)
}

// runServer helper
func (s *KeelTestSuite) runServer() {
	s.done = make(chan struct{})

	go func(done chan struct{}) {
		defer close(done)

		s.svr.Run()
	}(s.done)

	s.Require().Eventually(func() bool {
		statusCode, _, err := s.httpGet(s.url("/ok"))
		return err == nil && statusCode == http.StatusOK
	}, 5*time.Second, 10*time.Millisecond, "server did not start")
}

// waitServerStopped helper
func (s *KeelTestSuite) waitServerStopped() {
	if s.done == nil {
		return
	}

	select {
	case <-s.done:
	case <-time.After(10 * time.Second):
		s.FailNow("server did not stop")
	}
}

// url helper
func (s *KeelTestSuite) url(path string) string {
	return "http://" + s.addr + path
}

// httpGet helper
func (s *KeelTestSuite) httpGet(url string) (int, string, error) {
	if req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, url, nil); err != nil {
		return 0, "", err
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return 0, "", err
	} else if body, err := io.ReadAll(resp.Body); err != nil {
		return 0, "", err
	} else if err := resp.Body.Close(); err != nil {
		return 0, "", err
	} else {
		return resp.StatusCode, string(bytes.TrimSpace(body)), nil
	}
}

// httpPut helper
func (s *KeelTestSuite) httpPut(url, data string) (int, string, error) {
	if req, err := http.NewRequestWithContext(s.T().Context(), http.MethodPut, url, strings.NewReader(data)); err != nil {
		return 0, "", err
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return 0, "", err
	} else if body, err := io.ReadAll(resp.Body); err != nil {
		return 0, "", err
	} else if err := resp.Body.Close(); err != nil {
		return 0, "", err
	} else {
		return resp.StatusCode, string(bytes.TrimSpace(body)), nil
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKeelTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(KeelTestSuite))
}

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
