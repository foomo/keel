package roundtripware

import (
	"net/http"

	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	RequestIDOptions struct {
		Header string
	}
	RequestIDOption    func(*RequestIDOptions)
	RequestIDGenerator func() string
)

// GetDefaultRequestIDOptions returns the default options
func GetDefaultRequestIDOptions() RequestIDOptions {
	return RequestIDOptions{
		Header: "X-Request-ID",
	}
}

// RequestIDWithHeader middleware option
func RequestIDWithHeader(v string) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Header = v
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
				if value, ok := keelhttpcontext.GetRequestID(r.Context()); ok && value != "" {
					r.Header.Set(o.Header, value)
				}
			}
			return next(r)
		}
	}
}
