package middleware

import (
	"net/http"

	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/provider"
)

type (
	RequestIDOptions struct {
		Provider          provider.RequestID
		RequestHeader     []string
		ResponseHeader    string
		SetRequestHeader  bool
		SetResponseHeader bool
		SetContext        bool
	}
	RequestIDOption func(*RequestIDOptions)
)

// GetDefaultRequestIDOptions returns the default options
func GetDefaultRequestIDOptions() RequestIDOptions {
	return RequestIDOptions{
		Provider:          provider.DefaultRequestID,
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
func RequestIDWithGenerator(v provider.RequestID) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Provider = v
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
				requestID = opts.Provider()
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
