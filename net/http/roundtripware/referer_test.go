package roundtripware_test

import (
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

func TestReferer(t *testing.T) {
	var testReferer string

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testReferer = r.Header.Get(keelhttp.HeaderXReferer)
		assert.Empty(t, testReferer)
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Referer(),
		),
	)

	// create request
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testReferer, req.Header.Get(keelhttp.HeaderXReferer))
}

func TestReferer_Context(t *testing.T) {
	testReferer := "https://foomo.org/"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testReferer, r.Header.Get(keelhttp.HeaderXReferer))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Referer(),
		),
	)

	// set request id on context
	ctx := keelhttpcontext.SetReferer(t.Context(), testReferer)

	// create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testReferer, req.Header.Get(keelhttp.HeaderXReferer))
}

func TestReferer_WithHeader(t *testing.T) {
	testReferer := "https://foomo.org/"
	testHeader := "X-Custom-Header"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testReferer, r.Header.Get(testHeader))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Referer(
				roundtripware.RefererWithHeader(testHeader),
			),
		),
	)

	// set request id on context
	ctx := keelhttpcontext.SetReferer(t.Context(), testReferer)

	// create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
	require.NoError(t, err)

	// do request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// validate
	assert.Equal(t, testReferer, req.Header.Get(testHeader))
}
