package keel_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type KeelTestSuite struct {
	suite.Suite
	l      *zap.Logger
	svr    *keel.Server
	mux    *http.ServeMux
	cancel context.CancelFunc
}

// SetupSuite hook
func (s *KeelTestSuite) SetupSuite() {
	s.l = zaptest.NewLogger(s.T())
}

// BeforeTest hook
func (s *KeelTestSuite) BeforeTest(suiteName, testName string) {
	s.l = zaptest.NewLogger(s.T())
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s.mux.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 1)
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
	time.Sleep(time.Second * 3)
}

// TearDownSuite hook
func (s *KeelTestSuite) TearDownSuite() {}

func (s *KeelTestSuite) TestServiceHTTP() {
	s.svr.AddServices(
		service.NewHTTP(s.l, "test", "localhost:55000", s.mux),
	)

	s.runServer()

	if statusCode, _, err := s.httpGet("http://localhost:55000/ok"); s.NoError(err) {
		s.Equal(http.StatusOK, statusCode)
	}
}

func (s *KeelTestSuite) TestServiceHTTPZap() {
	s.svr.AddServices(
		service.NewHTTPZap(s.l, "zap", "localhost:9100", "/log"),
		service.NewHTTP(s.l, "test", "localhost:55000", s.mux),
	)

	s.runServer()

	s.Run("default", func() {
		if statusCode, body, err := s.httpGet("http://localhost:9100/log"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"info","disableCaller":true,"disableStacktrace":true}`, body)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/info"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/debug"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("set debug level", func() {
		if statusCode, body, err := s.httpPut("http://localhost:9100/log", `{"level":"debug"}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"debug","disableCaller":true,"disableStacktrace":true}`, body)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/info"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/debug"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("enable caller", func() {
		if statusCode, body, err := s.httpPut("http://localhost:9100/log", `{"disableCaller":false}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"debug","disableCaller":false,"disableStacktrace":true}`, body)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/error"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("enable stacktrace", func() {
		if statusCode, body, err := s.httpPut("http://localhost:9100/log", `{"disableStacktrace":false}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.JSONEq(`{"level":"debug","disableCaller":false,"disableStacktrace":false}`, body)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/error"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})
}

func (s *KeelTestSuite) TestGraceful() {
	s.svr.AddServices(
		service.NewHTTP(s.l, "test", "localhost:55000", s.mux),
	)

	s.runServer()

	{ // check that we're up
		if statusCode, _, err := s.httpGet("http://localhost:55000/ok"); s.NoError(err) {
			s.l.Info("received response from /ok")
			s.Equal(http.StatusOK, statusCode)
		}
	}

	{ // start long running call in separate process
		waitChan := make(chan string)
		go func(waitChan chan string) {
			waitChan <- "ok"
			s.l.Info("rending request to /sleep")
			if statusCode, _, err := s.httpGet("http://localhost:55000/sleep"); s.NoError(err) {
				s.l.Info("received response from /sleep")
				s.Equal(http.StatusOK, statusCode)
			}
		}(waitChan)
		s.l.Info("waiting for ")
		<-waitChan
	}

	{
		waitChan := make(chan string)
		go func(waitChan chan string) {
			waitChan <- "ok"
			time.Sleep(time.Second)
			if s.NoError(syscall.Kill(syscall.Getpid(), syscall.SIGINT)) {
				s.l.Info("killed myself")
			}
		}(waitChan)
		<-waitChan
	}

	time.Sleep(time.Second * 3)

	{ // check that server is down
		_, _, err := s.httpGet("http://localhost:55000/ok")
		s.Require().Error(err)
	}

	s.l.Info("done")
}

// runServer helper
func (s *KeelTestSuite) runServer() {
	l := s.svr.Logger()
	waitChan := make(chan string)
	go func(waitChan chan string) {
		waitChan <- "finished"
		s.svr.Run()
	}(waitChan)
	l.Debug("waiting for server process to start")
	<-waitChan
	time.Sleep(time.Second)
	l.Debug("continuing test")
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
	suite.Run(t, new(KeelTestSuite))
}
