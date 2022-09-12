package roundtripware

import (
	"net/http"

	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	SessionIDOptions struct {
		Header string
	}
	SessionIDOption    func(*SessionIDOptions)
	SessionIDGenerator func() string
)

// GetDefaultSessionIDOptions returns the default options
func GetDefaultSessionIDOptions() SessionIDOptions {
	return SessionIDOptions{
		Header: "X-Session-ID",
	}
}

// SessionIDWithHeader middleware option
func SessionIDWithHeader(v string) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Header = v
	}
}

// SessionID returns a RoundTripper which prints out the request & response object
func SessionID(opts ...SessionIDOption) RoundTripware {
	o := GetDefaultSessionIDOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			if value := r.Header.Get(o.Header); value == "" {
				if value, ok := keelhttpcontext.GetSessionID(r.Context()); ok && value != "" {
					r.Header.Set(o.Header, value)
				}
			}
			return next(r)
		}
	}
}
