package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foomo/keel/net/http/middleware"
	"go.uber.org/zap"
)

func BenchmarkCompose(b *testing.B) {
	noop := func(_ *zap.Logger, _ string, next http.Handler) http.Handler {
		return next
	}

	l := zap.NewNop()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	lengths := []int{1, 5, 10, 20}
	for _, n := range lengths {
		middlewares := make([]middleware.Middleware, n)
		for i := range n {
			middlewares[i] = noop
		}

		b.Run(fmt.Sprintf("chain-%d", n), func(b *testing.B) {
			for b.Loop() {
				middleware.Compose(l, "bench", handler, middlewares...)
			}
		})
	}
}

func BenchmarkComposeServe(b *testing.B) {
	passthrough := func(_ *zap.Logger, _ string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	l := zap.NewNop()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	lengths := []int{1, 5, 10, 20}
	for _, n := range lengths {
		middlewares := make([]middleware.Middleware, n)
		for i := range n {
			middlewares[i] = passthrough
		}

		composed := middleware.Compose(l, "bench", handler, middlewares...)
		w := httptest.NewRecorder()

		b.Run(fmt.Sprintf("serve-%d", n), func(b *testing.B) {
			for b.Loop() {
				composed.ServeHTTP(w, r)
			}
		})
	}
}
