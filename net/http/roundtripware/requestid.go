package roundtripware

import (
	"net/http"

	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/provider"
)

type (
	RequestIDOptions struct {
		Header    string
		Provider  provider.RequestID
		SetHeader bool
	}
	RequestIDOption func(*RequestIDOptions)
)

// GetDefaultRequestIDOptions returns the default options
func GetDefaultRequestIDOptions() RequestIDOptions {
	return RequestIDOptions{
		Header:   "X-Request-ID",
		Provider: provider.DefaultRequestID,
	}
}

// RequestIDWithHeader middleware option
func RequestIDWithHeader(v string) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Header = v
	}
}

// RequestIDWithProvider middleware option
func RequestIDWithProvider(v provider.RequestID) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Provider = v
	}
}

// RequestID returns a RoundTripper which prints out the request & response object
func RequestID(opts ...RequestIDOption) RoundTripware {
	o := GetDefaultRequestIDOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			if value := r.Header.Get(o.Header); value == "" {
				var requestID string
				if value, ok := keelhttpcontext.GetRequestID(r.Context()); ok && value != "" {
					requestID = value
				}
				if requestID == "" {
					requestID = o.Provider()
				}
				if requestID != "" {
					r.Header.Set(o.Header, requestID)
				}
			}
			return next(r)
		}
	}
}
