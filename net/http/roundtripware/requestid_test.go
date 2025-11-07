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
	t.Parallel()

	tests := []struct {
		name              string
		setupContext      func(t *testing.T) context.Context
		requestIDOptions  []roundtripware.RequestIDOption
		expectedHeader    string
		expectedRequestID string
		serverAssertions  func(t *testing.T, r *http.Request, capturedID *string)
		requestAssertions func(t *testing.T, req *http.Request, capturedID string)
	}{
		{
			name: "default behavior",
			setupContext: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			requestIDOptions: nil,
			expectedHeader:   keelhttp.HeaderXRequestID,
			serverAssertions: func(t *testing.T, r *http.Request, capturedID *string) {
				t.Helper()

				*capturedID = r.Header.Get(keelhttp.HeaderXRequestID)
				assert.NotEmpty(t, *capturedID)
			},
			requestAssertions: func(t *testing.T, req *http.Request, capturedID string) {
				t.Helper()
				assert.Equal(t, capturedID, req.Header.Get(keelhttp.HeaderXRequestID))
			},
		},
		{
			name: "with context",
			setupContext: func(t *testing.T) context.Context {
				t.Helper()
				return keelhttpcontext.SetRequestID(t.Context(), "123456")
			},
			requestIDOptions:  nil,
			expectedHeader:    keelhttp.HeaderXRequestID,
			expectedRequestID: "123456",
			serverAssertions: func(t *testing.T, r *http.Request, capturedID *string) {
				t.Helper()

				*capturedID = r.Header.Get(keelhttp.HeaderXRequestID)
				assert.Equal(t, "123456", *capturedID)
			},
			requestAssertions: func(t *testing.T, req *http.Request, capturedID string) {
				t.Helper()
				assert.Equal(t, "123456", req.Header.Get(keelhttp.HeaderXRequestID))
			},
		},
		{
			name: "with custom provider",
			setupContext: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			requestIDOptions: []roundtripware.RequestIDOption{
				roundtripware.RequestIDWithProvider(func() string {
					return "123456"
				}),
			},
			expectedHeader:    keelhttp.HeaderXRequestID,
			expectedRequestID: "123456",
			serverAssertions: func(t *testing.T, r *http.Request, capturedID *string) {
				t.Helper()

				*capturedID = r.Header.Get(keelhttp.HeaderXRequestID)
				assert.Equal(t, "123456", *capturedID)
			},
			requestAssertions: func(t *testing.T, req *http.Request, capturedID string) {
				t.Helper()
				assert.Equal(t, "123456", req.Header.Get(keelhttp.HeaderXRequestID))
			},
		},
		{
			name: "with custom header",
			setupContext: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			requestIDOptions: []roundtripware.RequestIDOption{
				roundtripware.RequestIDWithHeader("X-Custom-Header"),
			},
			expectedHeader: "X-Custom-Header",
			serverAssertions: func(t *testing.T, r *http.Request, capturedID *string) {
				t.Helper()

				*capturedID = r.Header.Get("X-Custom-Header")
				assert.NotEmpty(t, *capturedID)
			},
			requestAssertions: func(t *testing.T, req *http.Request, capturedID string) {
				t.Helper()
				assert.Equal(t, capturedID, req.Header.Get("X-Custom-Header"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequestID string

			l := zaptest.NewLogger(t)

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.serverAssertions(t, r, &capturedRequestID)
				w.WriteHeader(http.StatusOK)
			}))
			defer svr.Close()

			client := keelhttp.NewHTTPClient(
				keelhttp.HTTPClientWithRoundTripware(l,
					roundtripware.RequestID(tt.requestIDOptions...),
				),
			)

			ctx := tt.setupContext(t)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			tt.requestAssertions(t, req, capturedRequestID)
		})
	}
}
