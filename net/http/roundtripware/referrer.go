package roundtripware

import (
	"net/http"

	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	ReferrerOptions struct {
		Header string
	}
	ReferrerOption    func(*ReferrerOptions)
	ReferrerGenerator func() string
)

// GetDefaultReferrerOptions returns the default options
func GetDefaultReferrerOptions() ReferrerOptions {
	return ReferrerOptions{
		Header: "X-Referrer",
	}
}

// ReferrerWithHeader middleware option
func ReferrerWithHeader(v string) ReferrerOption {
	return func(o *ReferrerOptions) {
		o.Header = v
	}
}

// Referrer returns a RoundTripper which prints out the request & response object
func Referrer(opts ...ReferrerOption) RoundTripware {
	o := GetDefaultReferrerOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			if value := r.Header.Get(o.Header); value == "" {
				if value, ok := keelhttpcontext.GetReferrer(r.Context()); ok && value != "" {
					r.Header.Set(o.Header, value)
				}
			}
			return next(r)
		}
	}
}
