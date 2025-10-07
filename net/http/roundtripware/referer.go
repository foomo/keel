package roundtripware

import (
	"net/http"
	"strings"

	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	RefererOptions struct {
		Header string
	}
	RefererOption func(*RefererOptions)
)

// GetDefaultRefererOptions returns the default options
func GetDefaultRefererOptions() RefererOptions {
	return RefererOptions{
		Header: "X-Referer",
	}
}

// RefererWithHeader middleware option
func RefererWithHeader(v string) RefererOption {
	return func(o *RefererOptions) {
		o.Header = v
	}
}

// Referer returns a RoundTripper which prints out the request & response object
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
