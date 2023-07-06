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

func TestReferrer(t *testing.T) {
	var testReferrer string

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testReferrer = r.Header.Get(keelhttp.HeaderXReferrer)
		assert.Empty(t, testReferrer)
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Referrer(),
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
	assert.Equal(t, testReferrer, req.Header.Get(keelhttp.HeaderXReferrer))
}

func TestReferrer_Context(t *testing.T) {
	testReferrer := "https://foomo.org/"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testReferrer, r.Header.Get(keelhttp.HeaderXReferrer))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Referrer(),
		),
	)

	// set request id on context
	ctx := keelhttpcontext.SetReferrer(context.Background(), testReferrer)

	// create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testReferrer, req.Header.Get(keelhttp.HeaderXReferrer))
}

func TestReferrer_WithHeader(t *testing.T) {
	testReferrer := "https://foomo.org/"
	testHeader := "X-Custom-Header"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testReferrer, r.Header.Get(testHeader))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Referrer(
				roundtripware.ReferrerWithHeader(testHeader),
			),
		),
	)

	// set request id on context
	ctx := keelhttpcontext.SetReferrer(context.Background(), testReferrer)

	// create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testReferrer, req.Header.Get(testHeader))
}
