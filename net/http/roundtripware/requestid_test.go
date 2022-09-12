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
	l := zaptest.NewLogger(t)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(),
		),
	)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)
	assert.Empty(t, req.Header.Get(keelhttp.HeaderXRequestID))

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEmpty(t, req.Header.Get(keelhttp.HeaderXRequestID))
}

func TestRequestID_Context(t *testing.T) {
	l := zaptest.NewLogger(t)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(),
		),
	)
	ctx := keelhttpcontext.SetRequestID(context.Background(), "123456")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
	require.NoError(t, err)
	assert.Empty(t, req.Header.Get(keelhttp.HeaderXRequestID))

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "123456", req.Header.Get(keelhttp.HeaderXRequestID))
}
