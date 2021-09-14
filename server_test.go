package keel_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
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
		log.Logger().Info("logging info")
	})
	s.mux.HandleFunc("/log/debug", func(w http.ResponseWriter, r *http.Request) {
		log.Logger().Debug("logging debug")
	})
	s.mux.HandleFunc("/log/warn", func(w http.ResponseWriter, r *http.Request) {
		log.Logger().Warn("logging warn")
	})
	s.mux.HandleFunc("/log/error", func(w http.ResponseWriter, r *http.Request) {
		log.Logger().Error("logging error")
	})

	ctx, cancel := context.WithCancel(context.Background())
	s.svr = keel.NewServer(keel.WithContext(ctx))
	s.cancel = cancel
}

// AfterTest hook
func (s *KeelTestSuite) AfterTest(suiteName, testName string) {
	s.cancel()
}

// TearDownSuite hook
func (s *KeelTestSuite) TearDownSuite() {}

func (s *KeelTestSuite) TestServiceHTTP() {
	if os.Getenv("CI") != "" {
		s.T().Skip()
	}

	s.svr.AddServices(
		keel.NewServiceHTTP(log.Logger(), "test", ":55000", s.mux),
	)

	go s.svr.Run()

	if statusCode, _, err := s.httpGet("http://localhost:55000/ok"); s.NoError(err) {
		s.Equal(http.StatusOK, statusCode)
	}
}

func (s *KeelTestSuite) TestServiceHTTPZap() {
	if os.Getenv("CI") != "" {
		s.T().Skip()
	}

	s.svr.AddServices(
		keel.NewServiceHTTPZap(s.l, "zap", ":9100", "/log"),
		keel.NewServiceHTTP(s.l, "test", ":55000", s.mux),
	)

	go s.svr.Run()

	s.Run("default", func() {
		if statusCode, body, err := s.httpGet("http://localhost:9100/log"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.Equal(body, `{"level":"info","disableCaller":true,"disableStacktrace":true}`)
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
			s.Equal(body, `{"level":"debug","disableCaller":true,"disableStacktrace":true}`)
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
			s.Equal(body, `{"level":"debug","disableCaller":false,"disableStacktrace":true}`)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/error"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})

	s.Run("enable stacktrace", func() {
		if statusCode, body, err := s.httpPut("http://localhost:9100/log", `{"disableStacktrace":false}`); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
			s.Equal(body, `{"level":"debug","disableCaller":false,"disableStacktrace":false}`)
		}
		if statusCode, _, err := s.httpGet("http://localhost:55000/log/error"); s.NoError(err) {
			s.Equal(http.StatusOK, statusCode)
		}
	})
}

func (s *KeelTestSuite) TestGraceful() {
	if os.Getenv("CI") != "" {
		s.T().Skip()
	}

	s.svr.AddServices(
		keel.NewServiceHTTP(log.Logger(), "test", ":55000", s.mux),
	)

	go s.svr.Run()

	if statusCode, _, err := s.httpGet("http://localhost:55000/ok"); s.NoError(err) {
		s.l.Info("receiveds from ok")
		s.Equal(http.StatusOK, statusCode)
	}

	go func() {
		s.l.Info("calling sleep")
		if statusCode, _, err := s.httpGet("http://localhost:55000/sleep"); s.NoError(err) {
			s.l.Info("received resom sleep")
			s.Equal(http.StatusOK, statusCode)
		}
	}()

	go func() {
		time.Sleep(time.Second)
		s.cancel()
	}()

	time.Sleep(time.Second * 3)

	_, _, err := s.httpGet("http://localhost:55000/ok")
	s.Error(err)

	s.l.Info("done")
}

// httpGet helper
func (s *KeelTestSuite) httpGet(url string) (int, string, error) {
	if req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil); err != nil {
		return 0, "", err
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return 0, "", err
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return 0, "", err
	} else if err := resp.Body.Close(); err != nil {
		return 0, "", err
	} else {
		return resp.StatusCode, string(bytes.TrimSpace(body)), nil
	}
}

// httpPut helper
func (s *KeelTestSuite) httpPut(url, data string) (int, string, error) {
	if req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, url, strings.NewReader(data)); err != nil {
		return 0, "", err
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return 0, "", err
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
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
