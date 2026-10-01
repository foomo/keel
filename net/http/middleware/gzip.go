package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"

	stdhttp "github.com/foomo/gostandards/http"
	keelhttp "github.com/foomo/keel/net/http"
	httputils "github.com/foomo/keel/utils/net/http"
	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/gzip"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// GZipOptions configures the [GZip] middleware.
	GZipOptions struct {
		// CompressionLevel is the gzip level used to compress responses.
		CompressionLevel int
		// MinSize is the minimum response body size in bytes to compress.
		MinSize int
		// MaxDecompressedSize caps the number of decompressed request body bytes
		// accepted (guards against decompression bombs). Zero means unlimited.
		MaxDecompressedSize int64
	}
	// GZipOption configures [GZipOptions].
	GZipOption func(*GZipOptions)
)

// DefaultGZipOptions holds the defaults used by [GZip]: the default
// compression level, a minimum size of 1024 bytes and no decompression limit.
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

// GZipWithMinSize allows setting a minimum response body length to apply gzip compression (default: 1024 bytes).
func GZipWithMinSize(v int) GZipOption {
	return func(o *GZipOptions) {
		o.MinSize = v
	}
}

// GZipWithMaxDecompressedSize caps the decompressed request body size in bytes to guard
// against decompression bombs. Requests exceeding the limit are rejected with 413 (default: 0, unlimited).
func GZipWithMaxDecompressedSize(v int64) GZipOption {
	return func(o *GZipOptions) {
		o.MaxDecompressedSize = v
	}
}

// GZip returns a middleware that compresses responses for clients accepting
// gzip and decompresses request bodies sent with Content-Encoding: gzip. The
// Content-Encoding and Content-Length request headers are removed for
// decompressed bodies. Malformed gzip bodies are answered with 400 Bad Request.
func GZip(opts ...GZipOption) keelhttp.Middleware {
	options := DefaultGZipOptions

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return GZipWithOptions(options)
}

// GZipWithOptions is like [GZip] but takes fully populated options. The
// returned middleware panics when applied if the options are invalid.
func GZipWithOptions(opts GZipOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		pool := sync.Pool{
			New: func() any {
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
			if span.IsRecording() {
				span.AddEvent("GZip")
			}

			if !strings.EqualFold(r.Header.Get(stdhttp.HeaderContentEncoding.String()), stdhttp.EncodingGzip.String()) {
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
				// empty body: nothing to decompress, but strip the now-inaccurate encoding headers
				removeGZipRequestHeaders(r)
				next.ServeHTTP(w, r)

				return
			} else if err != nil {
				httputils.BadRequestServerError(l, w, r, errors.Wrap(err, "failed to reset gzip"))
				return
			}

			defer gr.Close()

			removeGZipRequestHeaders(r)

			if opts.MaxDecompressedSize > 0 {
				// read one byte past the limit so we can detect an overrun; memory stays bounded
				buf, err := io.ReadAll(io.LimitReader(gr, opts.MaxDecompressedSize+1))
				if err != nil {
					httputils.BadRequestServerError(l, w, r, errors.Wrap(err, "failed to read gzip body"))
					return
				}

				if int64(len(buf)) > opts.MaxDecompressedSize {
					httputils.RequestEntityTooLargeServerError(l, w, r, errors.New("gzip: decompressed size limit exceeded"))
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(buf))
			} else {
				r.Body = gr
			}

			next.ServeHTTP(w, r)
		}))
	}
}

// removeGZipRequestHeaders removes the request encoding/length headers that no longer describe
// the decompressed body handed to downstream handlers.
func removeGZipRequestHeaders(r *http.Request) {
	r.Header.Del(stdhttp.HeaderContentEncoding.String())
	r.Header.Del(stdhttp.HeaderContentLength.String())
	r.ContentLength = -1
}
