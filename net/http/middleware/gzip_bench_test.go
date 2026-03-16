package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	stdhttp "github.com/foomo/gostandards/http"
	"github.com/foomo/keel/net/http/middleware"
	"go.uber.org/zap"
)

func BenchmarkGZipDecompression(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			payload := strings.Repeat("A", sz.size)
			compressed := compressGzip(b, []byte(payload))

			handler := middleware.GZip()(zap.NewNop(), "bench", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body) //nolint:errcheck
				w.WriteHeader(http.StatusOK)
			}))

			for b.Loop() {
				r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressed))
				r.Header.Set(stdhttp.HeaderContentEncoding.String(), stdhttp.EncodingGzip.String())

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
			}
		})
	}
}

func compressGzip(b *testing.B, data []byte) []byte {
	b.Helper()

	var buf bytes.Buffer

	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		b.Fatal(err)
	}

	if err := gw.Close(); err != nil {
		b.Fatal(err)
	}

	return buf.Bytes()
}
