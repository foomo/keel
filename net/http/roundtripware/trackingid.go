package roundtripware

import (
	"net/http"

	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	TrackingIDOptions struct {
		Header string
	}
	TrackingIDOption    func(*TrackingIDOptions)
	TrackingIDGenerator func() string
)

// GetDefaultTrackingIDOptions returns the default options
func GetDefaultTrackingIDOptions() TrackingIDOptions {
	return TrackingIDOptions{
		Header: "X-Tracking-ID",
	}
}

// TrackingIDWithHeader middleware option
func TrackingIDWithHeader(v string) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Header = v
	}
}

// TrackingID returns a RoundTripper which prints out the request & response object
func TrackingID(opts ...TrackingIDOption) RoundTripware {
	o := GetDefaultTrackingIDOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			if value := r.Header.Get(o.Header); value == "" {
				if value, ok := keelhttpcontext.GetTrackingID(r.Context()); ok && value != "" {
					r.Header.Set(o.Header, value)
				}
			}
			return next(r)
		}
	}
}
