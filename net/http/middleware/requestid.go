package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	RequestIDOptions struct {
		Generator         RequestIDGenerator
		RequestHeader     []string
		ResponseHeader    string
		SetRequestHeader  bool
		SetResponseHeader bool
		SetContext        bool
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
		RequestHeader:     []string{"X-Request-ID", "Cf-Ray"},
		ResponseHeader:    "X-Request-ID",
		SetRequestHeader:  true,
		SetResponseHeader: true,
		SetContext:        true,
	}
}

// RequestIDWithRequestHeader middleware option
func RequestIDWithRequestHeader(v ...string) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.RequestHeader = append(o.RequestHeader, v...)
	}
}

// RequestIDWithResponseHeader middleware option
func RequestIDWithResponseHeader(v string) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.ResponseHeader = v
	}
}

// RequestIDWithSetRequestHeader middleware option
func RequestIDWithSetRequestHeader(v bool) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.SetRequestHeader = v
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

// RequestIDWithSetContext middleware option
func RequestIDWithSetContext(v bool) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.SetContext = v
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
			var requestID string
			for _, value := range opts.RequestHeader {
				if requestID = r.Header.Get(value); requestID != "" {
					break
				}
			}
			if requestID == "" {
				requestID = opts.Generator()
			}
			if requestID != "" && opts.SetContext {
				r = r.WithContext(keelhttpcontext.SetRequestID(r.Context(), requestID))
			}
			if requestID != "" && opts.SetRequestHeader {
				r.Header.Set(opts.ResponseHeader, requestID)
			}
			if requestID != "" && opts.SetResponseHeader {
				if value := w.Header().Get(opts.ResponseHeader); value == "" {
					w.Header().Add(opts.ResponseHeader, requestID)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
