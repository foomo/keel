package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	keelassert "github.com/foomo/keel/keeltest/assert"
	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/middleware"
	"go.uber.org/zap/zaptest"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name   string
		target string
		origin string
		method string
		with   middleware.Middleware
		expect func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder)
	}{
		{
			name:   "default",
			method: http.MethodGet,
			origin: "https://www.foomo.org",
			target: "https://www.foomo.org/",
			with:   middleware.CORS(),
			expect: func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder) {
				t.Helper()
				keelassert.InlineJSONEq(t, w.Header()) // INLINE: {"Access-Control-Allow-Origin":["*"],"Vary":["Origin"]}
			},
		},
		{
			name:   "allow methods",
			method: http.MethodOptions,
			origin: "https://www.foomo.org",
			target: "https://www.foomo.org/",
			with: middleware.CORS(
				middleware.CORSWithAllowMethods(http.MethodGet),
			),
			expect: func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder) {
				t.Helper()
				keelassert.InlineJSONEq(t, w.Header()) // INLINE: {"Access-Control-Allow-Methods":["GET"],"Access-Control-Allow-Origin":["*"],"Vary":["Origin","Access-Control-Request-Method","Access-Control-Request-Headers"]}
			},
		},
		{
			name:   "allow origins exact www.foomo.org",
			method: http.MethodGet,
			origin: "https://www.foomo.org",
			target: "https://www.foomo.org/",
			with: middleware.CORS(
				middleware.CORSWithAllowOrigins("https://www.foomo.org"),
			),
			expect: func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder) {
				t.Helper()
				keelassert.InlineJSONEq(t, w.Header()) // INLINE: {"Access-Control-Allow-Origin":["https://www.foomo.org"],"Vary":["Origin"]}
			},
		},
		{
			name:   "allow origins wildcard foomo.org",
			method: http.MethodGet,
			origin: "https://www.foomo.org",
			target: "https://www.foomo.org/",
			with: middleware.CORS(
				middleware.CORSWithAllowOrigins("*.foomo.org"),
			),
			expect: func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder) {
				t.Helper()
				keelassert.InlineJSONEq(t, w.Header()) // INLINE: {"Access-Control-Allow-Origin":["https://www.foomo.org"],"Vary":["Origin"]}
			},
		},
		{
			name:   "allow origins wildcard foomo.org with http",
			method: http.MethodGet,
			origin: "http://www.foomo.org",
			target: "http://www.foomo.org/",
			with: middleware.CORS(
				middleware.CORSWithAllowOrigins("https://*.foomo.org"),
			),
			expect: func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder) {
				t.Helper()
				keelassert.InlineJSONEq(t, w.Header()) // INLINE: {"Vary":["Origin"]}
			},
		},
		{
			name:   "allow origins wildcard foomo.org with https",
			method: http.MethodGet,
			origin: "https://www.foomo.org",
			target: "https://www.foomo.org/",
			with: middleware.CORS(
				middleware.CORSWithAllowOrigins("https://*.foomo.org"),
			),
			expect: func(t *testing.T, r *http.Request, w *httptest.ResponseRecorder) {
				t.Helper()
				keelassert.InlineJSONEq(t, w.Header()) // INLINE: {"Access-Control-Allow-Origin":["https://www.foomo.org"],"Vary":["Origin"]}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.with(zaptest.NewLogger(t), tt.name, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			r := httptest.NewRequest(tt.method, tt.target, nil)
			r.Header.Add(keelhttp.HeaderOrigin, tt.origin)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			tt.expect(t, r, w)
		})
	}
}
