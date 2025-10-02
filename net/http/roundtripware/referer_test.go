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

func TestReferer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		contextReferer  string
		headerToCheck   string
		refererOptions  []roundtripware.RefererOption
		setupContext    func(context.Context, string) context.Context
		serverAssertion func(*testing.T, *http.Request, string)
	}{
		{
			name:           "empty referer",
			contextReferer: "",
			headerToCheck:  keelhttp.HeaderXReferer,
			refererOptions: nil,
			setupContext:   nil,
			serverAssertion: func(t *testing.T, r *http.Request, expectedHeader string) {
				t.Helper()
				assert.Empty(t, r.Header.Get(expectedHeader))
			},
		},
		{
			name:           "referer from context",
			contextReferer: "https://foomo.org/",
			headerToCheck:  keelhttp.HeaderXReferer,
			refererOptions: nil,
			setupContext:   keelhttpcontext.SetReferer,
			serverAssertion: func(t *testing.T, r *http.Request, expectedHeader string) {
				t.Helper()
				assert.Equal(t, "https://foomo.org/", r.Header.Get(expectedHeader))
			},
		},
		{
			name:           "custom header",
			contextReferer: "https://foomo.org/",
			headerToCheck:  "X-Custom-Header",
			refererOptions: []roundtripware.RefererOption{
				roundtripware.RefererWithHeader("X-Custom-Header"),
			},
			setupContext: keelhttpcontext.SetReferer,
			serverAssertion: func(t *testing.T, r *http.Request, expectedHeader string) {
				t.Helper()
				assert.Equal(t, "https://foomo.org/", r.Header.Get(expectedHeader))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create logger
			l := zaptest.NewLogger(t)

			// create http server with handler
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.serverAssertion(t, r, tt.headerToCheck)
				w.WriteHeader(http.StatusOK)
			}))
			defer svr.Close()

			// create http client
			client := keelhttp.NewHTTPClient(
				keelhttp.HTTPClientWithRoundTripware(l,
					roundtripware.Referer(tt.refererOptions...),
				),
			)

			// setup context
			ctx := t.Context()
			if tt.setupContext != nil {
				ctx = tt.setupContext(ctx, tt.contextReferer)
			}

			// create request
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL, nil)
			require.NoError(t, err)

			// do request
			resp, err := client.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			// validate
			if tt.contextReferer != "" {
				assert.Equal(t, tt.contextReferer, req.Header.Get(tt.headerToCheck))
			}
		})
	}
}
