package roundtripware

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"sync"

	stdhttp "github.com/foomo/gostandards/http"
	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/gzip"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	GZipOptions struct {
		CompressionLevel int
		MinSize          int
	}
	GZipOption func(*GZipOptions)
)

// DefaultGZipOptions returns the default options
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

// GZip returns a RoundTripware which logs all requests
func GZip(opts ...GZipOption) RoundTripware {
	o := DefaultGZipOptions

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		pool := sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(nil)
			},
		}
		wrapper := gzhttp.Transport(RoundTripperFunc(next))

		return func(req *http.Request) (*http.Response, error) {
			// Check if the request has a body
			if req.Body != nil && req.Header.Get(stdhttp.HeaderContentEncoding.String()) != stdhttp.EncodingGzip.String() && req.ContentLength >= int64(o.MinSize) {
				// Create a buffer to store the gzipped data
				var buf bytes.Buffer

				// Get a gzip writer from the pool
				gzipWriter, ok := pool.Get().(*gzip.Writer)
				if !ok {
					return nil, errors.New("gzip writer: not a gzip writer")
				}
				// Reset the gzip writer to use the buffer
				gzipWriter.Reset(&buf)
				// Return the gzip writer to the pool for reuse
				defer pool.Put(gzipWriter)

				// Copy the request body into the gzip writer
				_, err := io.Copy(gzipWriter, req.Body)
				if err != nil {
					return nil, errors.Wrap(err, "failed to copy body")
				}

				// Close the gzip writer to flush any remaining data
				if err := gzipWriter.Close(); err != nil {
					return nil, errors.Wrap(err, "failed to close gzip writer")
				}

				// Close the original request body
				if err := req.Body.Close(); err != nil {
					return nil, errors.Wrap(err, "failed to close request body")
				}

				// Replace the original body with the gzipped body
				req.Body = io.NopCloser(&buf)

				// Set Content-Encoding header to indicate gzip compression
				req.Header.Set(stdhttp.HeaderContentEncoding.String(), stdhttp.EncodingGzip.String())

				// Optional: Set the Content-Length header
				cotentLength := buf.Len()
				req.Header.Set(stdhttp.HeaderContentLength.String(), strconv.Itoa(cotentLength))
				req.ContentLength = int64(cotentLength)
			}

			return wrapper.RoundTrip(req)
		}
	}
}
