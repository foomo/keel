package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	keelhttp "github.com/foomo/keel/net/http"
)

// responseWriter wraps an [http.ResponseWriter] to record the status code and
// the number of body bytes written, and to optionally set the
// X-Response-Time header.
type responseWriter struct {
	http.ResponseWriter
	writeResponseTimeHeader bool
	wroteHeader             bool
	statusCode              int
	start                   time.Time
	size                    int
}

// WrapResponseWriter wraps w to record the status code and body size. If w is
// already wrapped it is returned as is.
func WrapResponseWriter(w http.ResponseWriter) *responseWriter {
	if wr, ok := w.(*responseWriter); ok {
		return wr
	}

	return &responseWriter{
		ResponseWriter: w,
		start:          time.Now(),
	}
}

// SetWriteResponseTimeHeader sets whether WriteHeader adds the
// X-Response-Time header with the microseconds elapsed since wrapping.
func (w *responseWriter) SetWriteResponseTimeHeader(write bool) {
	w.writeResponseTimeHeader = write
}

// Size returns the number of body bytes written.
func (w *responseWriter) Size() int {
	return w.size
}

// StatusCode returns the written status code, or 200 if none was written.
func (w *responseWriter) StatusCode() int {
	if !w.wroteHeader {
		return http.StatusOK
	}

	return w.statusCode
}

// Status returns StatusCode as a decimal string.
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

// WriteHeader records statusCode and forwards it. Only the first call takes
// effect; later calls are ignored.
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

// Write writes b, writing a 200 status first if no header was written, and
// adds the written bytes to the size.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	size, err := w.ResponseWriter.Write(b)
	w.size += size

	return size, err
}
