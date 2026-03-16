package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/middleware"
	"go.uber.org/zap"
)

func BenchmarkMatchSubdomain(b *testing.B) {
	cases := []struct {
		name    string
		domain  string
		pattern string
	}{
		{"exact", "https://www.foomo.org", "https://www.foomo.org"},
		{"wildcard-shallow", "https://www.foomo.org", "https://*.foomo.org"},
		{"wildcard-deep", "https://a.b.c.d.foomo.org", "https://*.foomo.org"},
		{"no-match", "https://www.example.com", "https://*.foomo.org"},
		{"scheme-mismatch", "http://www.foomo.org", "https://*.foomo.org"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				middleware.BenchMatchSubdomain(tc.domain, tc.pattern)
			}
		})
	}
}

func BenchmarkCORSOriginMatching(b *testing.B) {
	origins := []string{
		"https://app.foomo.org",
		"https://api.foomo.org",
		"https://*.example.com",
	}

	handler := middleware.CORSWithOptions(middleware.CORSOptions{
		AllowOrigins: origins,
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	})

	l := zap.NewNop()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := handler(l, "bench", next)

	cases := []struct {
		name   string
		origin string
		method string
	}{
		{"exact-match", "https://app.foomo.org", http.MethodGet},
		{"pattern-match", "https://sub.example.com", http.MethodGet},
		{"no-match", "https://evil.com", http.MethodGet},
		{"preflight-match", "https://app.foomo.org", http.MethodOptions},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			r := httptest.NewRequest(tc.method, "https://localhost/", nil)
			r.Header.Set(keelhttp.HeaderOrigin, tc.origin)

			w := httptest.NewRecorder()

			for b.Loop() {
				h.ServeHTTP(w, r)
				w.Body.Reset()
			}
		})
	}
}
