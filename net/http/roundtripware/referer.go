package roundtripware

import (
	"net/http"
	"strings"

	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	// RefererOptions configures the [Referer] RoundTripware.
	RefererOptions struct {
		// Header is the name of the request header to set.
		Header string
	}
	// RefererOption configures [RefererOptions].
	RefererOption func(*RefererOptions)
)

// GetDefaultRefererOptions returns the default options using the X-Referer
// header.
func GetDefaultRefererOptions() RefererOptions {
	return RefererOptions{
		Header: "X-Referer",
	}
}

// RefererWithHeader sets the request header name. Defaults to X-Referer.
func RefererWithHeader(v string) RefererOption {
	return func(o *RefererOptions) {
		o.Header = v
	}
}

// Referer returns a RoundTripware that sets the referer stored in the request
// context as request header, unless the header is already present.
func Referer(opts ...RefererOption) RoundTripware {
	o := GetDefaultRefererOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Referer")
			}

			if value := r.Header.Get(o.Header); value == "" {
				if value, ok := keelhttpcontext.GetReferer(r.Context()); ok && value != "" {
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
