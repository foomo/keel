package roundtripware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	keelhttp "github.com/foomo/keel/net/http"
	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/roundtripware"
)

func TestRequestID(t *testing.T) {
	var testRequestID string

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testRequestID = r.Header.Get(keelhttp.HeaderXRequestID)
		assert.NotEmpty(t, testRequestID)
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testRequestID, req.Header.Get(keelhttp.HeaderXRequestID))
}

func TestRequestID_Context(t *testing.T) {
	testRequestID := "123456"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testRequestID, r.Header.Get(keelhttp.HeaderXRequestID))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(),
		),
	)

	// set request id on context
	ctx := keelhttpcontext.SetRequestID(context.Background(), testRequestID)

	// create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testRequestID, req.Header.Get(keelhttp.HeaderXRequestID))
}

func TestRequestID_WithProvider(t *testing.T) {
	testRequestID := "123456"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testRequestID, r.Header.Get(keelhttp.HeaderXRequestID))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(
				roundtripware.RequestIDWithProvider(func() string {
					return testRequestID
				}),
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

	// validate
	assert.Equal(t, testRequestID, req.Header.Get(keelhttp.HeaderXRequestID))
}

func TestRequestID_WithHeader(t *testing.T) {
	var testRequestID string
	testHeader := "X-Custom-Header"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testRequestID = r.Header.Get(testHeader)
		assert.NotEmpty(t, testRequestID)
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(
				roundtripware.RequestIDWithHeader(testHeader),
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

	// validate
	assert.Equal(t, testRequestID, req.Header.Get(testHeader))
}
