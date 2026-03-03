package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/foomo/keel/net/http/middleware"
)

func BenchmarkWrapResponseWriter(b *testing.B) {
	b.Run("new", func(b *testing.B) {
		w := httptest.NewRecorder()

		for b.Loop() {
			middleware.WrapResponseWriter(w)
		}
	})

	b.Run("already-wrapped", func(b *testing.B) {
		w := middleware.WrapResponseWriter(httptest.NewRecorder())

		for b.Loop() {
			middleware.WrapResponseWriter(w)
		}
	})
}

func BenchmarkResponseWriterWrite(b *testing.B) {
	b.Run("single-large", func(b *testing.B) {
		data := make([]byte, 4096)
		w := middleware.WrapResponseWriter(httptest.NewRecorder())

		for b.Loop() {
			_, _ = w.Write(data)
		}
	})

	b.Run("many-small", func(b *testing.B) {
		data := make([]byte, 64)
		w := middleware.WrapResponseWriter(httptest.NewRecorder())

		for b.Loop() {
			for range 64 {
				_, _ = w.Write(data)
			}
		}
	})
}
