package roundtripware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/roundtripware"
)

func TestLogger(t *testing.T) {
	// create logger & validate output
	l := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
		assert.Equal(t, zapcore.InfoLevel, entry.Level)
		assert.Equal(t, "sent request", entry.Message)
		return nil
	})))

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Logger(),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}

func TestLogger_WithMessage(t *testing.T) {
	testMessage := "my message"

	// create logger & validate output
	l := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
		assert.Equal(t, testMessage, entry.Message)
		return nil
	})))

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Logger(
				roundtripware.LoggerWithMessage(testMessage),
			),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}

func TestLogger_WithErrorMessage(t *testing.T) {
	testErrorMessage := "my error message"

	// create logger & validate output
	l := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
		assert.Equal(t, testErrorMessage, entry.Message)
		return nil
	})))

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			// trigger an internal error
			func(l *zap.Logger, next roundtripware.Handler) roundtripware.Handler {
				return func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("something went wrong")
				}
			},
			roundtripware.Logger(
				roundtripware.LoggerWithErrorMessage(testErrorMessage),
			),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.Nil(t, resp)
	require.Error(t, err)
	if resp != nil {
		defer resp.Body.Close()
	}
}

func TestLogger_WithMinWarnCode(t *testing.T) {
	// create logger & validate output
	l := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
		assert.Equal(t, zapcore.WarnLevel, entry.Level)
		return nil
	})))

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Logger(
				roundtripware.LoggerWithMinWarnCode(200),
			),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}

func TestLogger_WithMinErrorCode(t *testing.T) {
	// create logger & validate output
	l := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
		assert.Equal(t, zapcore.ErrorLevel, entry.Level)
		return nil
	})))

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Logger(
				roundtripware.LoggerWithMinErrorCode(200),
			),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}
