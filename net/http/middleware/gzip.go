package middleware

import (
	"io"
	"net/http"
	"sync"

	stdhttp "github.com/foomo/gostandards/http"
	httputils "github.com/foomo/keel/utils/net/http"
	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/gzip"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	GZipOptions struct {
		CompressionLevel int
		MinSize          int
	}
	GZipOption func(*GZipOptions)
)

var DefaultGZipOptions = GZipOptions{
	CompressionLevel: gzip.DefaultCompression,
	MinSize:          1024,
}

// GZipWithLevel allows setting a specific compression level for gzip (default: gzip.DefaultCompression).
func GZipWithLevel(v int) GZipOption {
	return func(o *GZipOptions) {
		o.CompressionLevel = v
	}
}

// GZipWithMinSize allows setting a minimum response body length to apply gzip compression (default: 1400 bytes).
func GZipWithMinSize(v int) GZipOption {
	return func(o *GZipOptions) {
		o.MinSize = v
	}
}

// GZip middleware
func GZip(opts ...GZipOption) Middleware {
	options := DefaultGZipOptions

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return GZipWithOptions(options)
}

// GZipWithOptions middleware
func GZipWithOptions(opts GZipOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		pool := sync.Pool{
			New: func() interface{} {
				return new(gzip.Reader)
			},
		}

		wrapper, err := gzhttp.NewWrapper(
			gzhttp.CompressionLevel(opts.CompressionLevel),
			gzhttp.MinSize(opts.MinSize),
		)
		if err != nil {
			panic(err)
		}

		return wrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			span.AddEvent("GZip")

			if r.Header.Get(stdhttp.HeaderContentEncoding.String()) != stdhttp.EncodingGzip.String() {
				next.ServeHTTP(w, r)
				return
			}

			gr, ok := pool.Get().(*gzip.Reader)
			if !ok {
				httputils.InternalServerError(l, w, r, errors.New("failed to retrieve gzip pool"))
				return
			}
			defer pool.Put(gr)

			b := r.Body
			defer b.Close()

			if err := gr.Reset(b); errors.Is(err, io.EOF) {
				next.ServeHTTP(w, r)
				return
			} else if err != nil {
				httputils.BadRequestServerError(l, w, r, errors.New("failed to reset gzip"))
				return
			}

			defer gr.Close()

			r.Header.Del(stdhttp.HeaderContentEncoding.String())

			r.Body = gr

			next.ServeHTTP(w, r)
		}))
	}
}
