package roundtripware

import (
	"net/http"
	"strings"

	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	// TrackingIDOptions configures the [TrackingID] RoundTripware.
	TrackingIDOptions struct {
		// Header is the name of the request header to set.
		Header string
	}
	// TrackingIDOption configures [TrackingIDOptions].
	TrackingIDOption func(*TrackingIDOptions)
	// TrackingIDGenerator returns a new tracking ID. It is not used by this
	// package.
	TrackingIDGenerator func() string
)

// GetDefaultTrackingIDOptions returns the default options using the
// X-Tracking-ID header.
func GetDefaultTrackingIDOptions() TrackingIDOptions {
	return TrackingIDOptions{
		Header: "X-Tracking-ID",
	}
}

// TrackingIDWithHeader sets the request header name. Defaults to
// X-Tracking-ID.
func TrackingIDWithHeader(v string) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Header = v
	}
}

// TrackingID returns a RoundTripware that sets the tracking ID stored in the
// request context as request header, unless the header is already present.
func TrackingID(opts ...TrackingIDOption) RoundTripware {
	o := GetDefaultTrackingIDOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("TrackingID")
			}

			if value := r.Header.Get(o.Header); value == "" {
				if value, ok := keelhttpcontext.GetTrackingID(r.Context()); ok && value != "" {
					if span.IsRecording() {
						span.SetAttributes(semconv.HTTPRequestHeader(strings.ToLower(o.Header), value))
					}

					r.Header.Set(o.Header, value)
				}
			}

			return next(r)
		}
	}
}
