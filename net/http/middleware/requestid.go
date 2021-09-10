package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	httputils "github.com/foomo/keel/net/http"
)

type (
	RequestIDOptions struct {
		Generator         RequestIDGenerator
		ResponseHeader    string
		SetResponseHeader bool
	}
	RequestIDOption    func(*RequestIDOptions)
	RequestIDGenerator func() string
)

// DefaultRequestIDGenerator function
func DefaultRequestIDGenerator() string {
	return uuid.New().String()
}

// GetDefaultRequestIDOptions returns the default options
func GetDefaultRequestIDOptions() RequestIDOptions {
	return RequestIDOptions{
		Generator:         DefaultRequestIDGenerator,
		ResponseHeader:    httputils.HeaderXRequestID,
		SetResponseHeader: false,
	}
}

// RequestIDWithResponseHeader middleware option
func RequestIDWithResponseHeader(v string) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.ResponseHeader = v
	}
}

// RequestIDWithSetResponseHeader middleware option
func RequestIDWithSetResponseHeader(v bool) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.SetResponseHeader = v
	}
}

// RequestIDWithGenerator middleware option
func RequestIDWithGenerator(v RequestIDGenerator) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Generator = v
	}
}

// RequestID middleware
func RequestID(opts ...RequestIDOption) Middleware {
	options := GetDefaultRequestIDOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return RequestIDWithOptions(options)
}

// RequestIDWithOptions middleware
func RequestIDWithOptions(opts RequestIDOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(opts.ResponseHeader)
			if requestID == "" {
				requestID = opts.Generator()
				r.Header.Set(opts.ResponseHeader, requestID)
			}
			if opts.SetResponseHeader {
				if value := w.Header().Get(opts.ResponseHeader); value == "" {
					w.Header().Add(opts.ResponseHeader, requestID)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
