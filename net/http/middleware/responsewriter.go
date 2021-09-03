package middleware

import (
	"fmt"
	"net/http"
)

// responseWriter is a wrapper that includes that http statusCode and size for logging
type responseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
	size        int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	if wr, ok := w.(*responseWriter); ok {
		return wr
	}
	return &responseWriter{
		ResponseWriter: w,
	}
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
	w.wroteHeader = true
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}
