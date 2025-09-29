package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	http2 "github.com/foomo/keel/net/http"
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

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.writeResponseTimeHeader {
		w.Header().Set(http2.HeaderXResponseTime, strconv.FormatInt(time.Since(w.start).Microseconds(), 10))
	}
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
	w.wroteHeader = true
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}
