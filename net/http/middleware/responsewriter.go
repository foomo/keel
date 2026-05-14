package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	keelhttp "github.com/foomo/keel/net/http"
)

// responseWriter is a wrapper that includes that http statusCode and size for logging
type responseWriter struct {
	http.ResponseWriter
	writeResponseTimeHeader bool
	wroteHeader             bool
	statusCode              int
	start                   time.Time
	size                    int
}

func WrapResponseWriter(w http.ResponseWriter) *responseWriter {
	if wr, ok := w.(*responseWriter); ok {
		return wr
	}

	return &responseWriter{
		ResponseWriter: w,
		start:          time.Now(),
	}
}

func (w *responseWriter) SetWriteResponseTimeHeader(write bool) {
	w.writeResponseTimeHeader = write
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) StatusCode() int {
	if !w.wroteHeader {
		return http.StatusOK
	}

	return w.statusCode
}

func (w *responseWriter) Status() string {
	return fmt.Sprintf("%d", w.StatusCode())
}

// Unwrap returns the underlying http.ResponseWriter, allowing
// http.ResponseController to discover optional interfaces (e.g., http.Flusher).
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Flush implements http.Flusher by delegating to the underlying writer if it
// supports flushing. This enables SSE (Server-Sent Events) and HTTP streaming.
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.statusCode = statusCode

	if w.writeResponseTimeHeader {
		w.Header().Set(keelhttp.HeaderXResponseTime, strconv.FormatInt(time.Since(w.start).Microseconds(), 10))
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	size, err := w.ResponseWriter.Write(b)
	w.size += size

	return size, err
}
