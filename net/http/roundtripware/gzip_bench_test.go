package roundtripware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/foomo/keel/net/http/roundtripware"
	"go.uber.org/zap"
)

func BenchmarkRoundtripGZipCompression(b *testing.B) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}))
	defer svr.Close()

	l := zap.NewNop()
	gzipRT := roundtripware.GZip()(l, http.DefaultTransport.RoundTrip)

	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
	}

	for _, sz := range sizes {
		payload := strings.Repeat("A", sz.size)

		b.Run(sz.name, func(b *testing.B) {
			for b.Loop() {
				req, err := http.NewRequestWithContext(b.Context(), http.MethodPost, svr.URL, strings.NewReader(payload))
				if err != nil {
					b.Fatal(err)
				}

				req.ContentLength = int64(sz.size)

				resp, err := gzipRT(req)
				if err != nil {
					b.Fatal(err)
				}

				_ = resp.Body.Close()
			}
		})
	}
}
