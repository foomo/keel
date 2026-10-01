package keel_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
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

// shutdownSignal is only registered by the signal trigger, so sending it does not
// stop servers of tests running in parallel.
const shutdownSignal = syscall.SIGUSR2

type KeelTestSuite struct {
	suite.Suite
	l       *zap.Logger
	svr     *keel.Server
	mux     *http.ServeMux
	addr    string
	zapAddr string
	done    chan struct{}
	cancel  context.CancelFunc
}

// BeforeTest hook
func (s *KeelTestSuite) BeforeTest(suiteName, testName string) {
	ports := testingx.FreePorts(s.T(), 2)
	s.addr = fmt.Sprintf("localhost:%d", ports[0])
	s.zapAddr = fmt.Sprintf("localhost:%d", ports[1])
	s.done = nil

	s.l = zaptest.NewLogger(s.T())
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
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

// TestServerGracefulShutdown asserts the graceful shutdown contract for every
// trigger: readiness fails and new connections are refused while in-flight
// requests complete with a live context, then Run returns.
func TestServerGracefulShutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []keel.Option
		trigger func(t *testing.T, svr *keel.Server, fail chan<- error)
	}{
		{
			name: "shutdown cancel",
			trigger: func(t *testing.T, svr *keel.Server, _ chan<- error) {
				t.Helper()
				svr.ShutdownCancel()()
			},
		},
		{
			name: "signal",
			opts: []keel.Option{keel.WithShutdownSignals(shutdownSignal)},
			trigger: func(t *testing.T, _ *keel.Server, _ chan<- error) {
				t.Helper()
				require.NoError(t, syscall.Kill(syscall.Getpid(), shutdownSignal))
			},
		},
		{
			name: "service failure",
			trigger: func(t *testing.T, _ *keel.Server, fail chan<- error) {
				t.Helper()

				fail <- errors.New("boom")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := zaptest.NewLogger(t)
			ports := testingx.FreePorts(t, 2)
			appURL := fmt.Sprintf("http://localhost:%d", ports[0])
			readinessURL := fmt.Sprintf("http://localhost:%d/healthz/readiness", ports[1])

			entered := make(chan struct{})
			release := make(chan struct{})
			handlerCtxErr := make(chan error, 1)

			mux := http.NewServeMux()
			mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
			mux.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
				close(entered)
				<-release

				handlerCtxErr <- r.Context().Err()
			})

			svr := keel.NewServer(append([]keel.Option{
				keel.WithContext(t.Context()),
				keel.WithLogger(l),
				keel.WithGracefulPeriod(5 * time.Second),
			}, tt.opts...)...)

			fail := make(chan error, 1)
			gate := make(chan struct{})
			gateCalls := make(chan closerCall, 1)

			// closers run in registration order: the gate holds the shutdown before
			// any service is closed, and healthz is added last so that readiness
			// stays observable while the app service drains
			svr.AddCloser(&testCloser{name: "gate", calls: gateCalls, gate: gate})
			svr.AddServices(
				service.NewHTTP(l, "app", fmt.Sprintf("localhost:%d", ports[0]), mux),
				service.NewGoRoutine(l, "worker", func(ctx context.Context, _ *zap.Logger) error {
					select {
					case err := <-fail:
						return err
					case <-ctx.Done():
						return nil
					}
				}),
				service.NewHealthz(l, "healthz", fmt.Sprintf("localhost:%d", ports[1]), "/healthz", svr.ProbesForTest()),
			)

			done := startServer(t, svr)

			// unblock on failure, so the server can drain on cleanup
			unblock := sync.OnceFunc(func() { close(release) })
			t.Cleanup(unblock)

			openGate := sync.OnceFunc(func() { close(gate) })
			t.Cleanup(openGate)

			requireEventuallyStatus(t, appURL+"/ok", http.StatusOK)
			requireEventuallyStatus(t, readinessURL, http.StatusOK)

			// start in-flight request
			inFlight := make(chan httpResult, 1)

			go func() { inFlight <- httpGet(t.Context(), appURL+"/block") }()

			receive(t, entered, "request did not reach the handler")

			tt.trigger(t, svr, fail)

			// no service is closed yet: readiness must fail before the listeners close
			receive(t, gateCalls, "shutdown did not start")
			requireEventuallyStatus(t, readinessURL, http.StatusServiceUnavailable)
			requireEventuallyStatus(t, appURL+"/ok", http.StatusOK)

			openGate()

			require.Eventually(t, func() bool {
				return httpGet(t.Context(), appURL+"/ok").err != nil
			}, 5*time.Second, 10*time.Millisecond, "new connections must be refused while draining")

			select {
			case <-done:
				t.Fatal("server stopped before the in-flight request completed")
			default:
			}

			unblock()

			require.NoError(t, receive(t, handlerCtxErr, "handler did not complete"), "in-flight request context must stay alive")

			res := receive(t, inFlight, "in-flight request did not complete")
			require.NoError(t, res.err)
			require.Equal(t, http.StatusOK, res.statusCode)

			receive(t, done, "server did not stop")
			require.ErrorIs(t, svr.Healthz(), keel.ErrServerNotRunning)
		})
	}
}

// TestServerShutdownClosers asserts closers run in registration order and a
// failing closer does not prevent the remaining ones from running.
func TestServerShutdownClosers(t *testing.T) {
	t.Parallel()

	calls := make(chan closerCall, 3)

	svr := keel.NewServer(
		keel.WithContext(t.Context()),
		keel.WithLogger(zaptest.NewLogger(t)),
	)
	svr.AddClosers(
		&testCloser{name: "first", calls: calls},
		&testCloser{name: "failing", calls: calls, err: errors.New("boom")},
		&testCloser{name: "last", calls: calls},
	)

	done := startServer(t, svr)

	svr.ShutdownCancel()()
	receive(t, done, "server did not stop")

	for _, name := range []string{"first", "failing", "last"} {
		require.Equal(t, name, receive(t, calls, "closer was not called").name)
	}
}

// TestServerGracefulPeriodExceeded asserts the graceful period bounds the
// shutdown: a hanging closer is given up on and the remaining closers still run.
func TestServerGracefulPeriodExceeded(t *testing.T) {
	t.Parallel()

	calls := make(chan closerCall, 2)

	svr := keel.NewServer(
		keel.WithContext(t.Context()),
		keel.WithLogger(zaptest.NewLogger(t)),
		keel.WithGracefulPeriod(100*time.Millisecond),
	)
	svr.AddClosers(
		&testCloser{name: "hanging", calls: calls, block: true},
		&testCloser{name: "last", calls: calls},
	)

	done := startServer(t, svr)

	svr.ShutdownCancel()()
	receive(t, done, "server did not stop after the graceful period")

	hanging := receive(t, calls, "hanging closer was not called")
	require.Equal(t, "hanging", hanging.name)
	require.ErrorIs(t, hanging.err, context.DeadlineExceeded)
	require.Equal(t, "last", receive(t, calls, "last closer was not called").name)
}

// TestServerContextCancel asserts cancelling the server context is a hard stop:
// in-flight requests see their context cancelled, but closers still get a live
// context bounded by the graceful period, e.g. to flush telemetry.
func TestServerContextCancel(t *testing.T) {
	t.Parallel()

	l := zaptest.NewLogger(t)
	addr := fmt.Sprintf("localhost:%d", testingx.FreePort(t))
	ctx, cancel := context.WithCancel(t.Context())

	entered := make(chan struct{})
	handlerCtxErr := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()

		handlerCtxErr <- r.Context().Err()
	})

	calls := make(chan closerCall, 1)

	svr := keel.NewServer(
		keel.WithContext(ctx),
		keel.WithLogger(l),
		keel.WithGracefulPeriod(5*time.Second),
	)
	svr.AddService(service.NewHTTP(l, "app", addr, mux))
	svr.AddCloser(&testCloser{name: "closer", calls: calls})

	done := startServer(t, svr)

	requireEventuallyStatus(t, "http://"+addr+"/ok", http.StatusOK)

	go func() { _ = httpGet(t.Context(), "http://"+addr+"/block") }()

	receive(t, entered, "request did not reach the handler")

	cancel()

	require.ErrorIs(t, receive(t, handlerCtxErr, "handler did not complete"), context.Canceled)

	call := receive(t, calls, "closer was not called")
	require.NoError(t, call.err, "closer must not receive a cancelled context")
	require.True(t, call.hasDeadline, "closer context must be bounded by the graceful period")

	receive(t, done, "server did not stop")
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

	done := startServer(t, svr)
	receive(t, done, "server kept running although a service could not bind its port")

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

	done := startServer(t, svr)
	receive(t, done, "server kept running although a service could not bind its port")

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
