package middleware

import (
	"net/http"
	"strings"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/context"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// RefererOptions configures the [Referer] middleware.
	RefererOptions struct {
		// RequestHeader lists the request headers checked in order; the first
		// non-empty value is used.
		RequestHeader []string
		// SetContext reports whether the referer is stored in the request
		// context.
		SetContext bool
	}
	// RefererOption configures [RefererOptions].
	RefererOption func(*RefererOptions)
)

// GetDefaultRefererOptions returns the default options, which read the
// X-Referer and Referer headers and set the context.
func GetDefaultRefererOptions() RefererOptions {
	return RefererOptions{
		RequestHeader: []string{"X-Referer", "Referer"},
		SetContext:    true,
	}
}

// RefererWithRequestHeader appends headers to [RefererOptions.RequestHeader].
func RefererWithRequestHeader(v ...string) RefererOption {
	return func(o *RefererOptions) {
		o.RequestHeader = append(o.RequestHeader, v...)
	}
}

// RefererWithSetContext sets [RefererOptions.SetContext]. Defaults to true.
func RefererWithSetContext(v bool) RefererOption {
	return func(o *RefererOptions) {
		o.SetContext = v
	}
}

// Referer returns a middleware that reads the referer from the first
// non-empty configured request header, records it on the span and stores it
// in the request context.
func Referer(opts ...RefererOption) keelhttp.Middleware {
	options := GetDefaultRefererOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return RefererWithOptions(options)
}

// RefererWithOptions is like [Referer] but takes fully populated options.
func RefererWithOptions(opts RefererOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Referer")
			}

			var (
				key     string
				referer string
			)
			for _, value := range opts.RequestHeader {
				if referer = r.Header.Get(value); referer != "" {
					key = value
					break
				}
			}

			if span.IsRecording() && referer != "" {
				span.SetAttributes(semconv.HTTPRequestHeader(strings.ToLower(key), referer))
			}

			if referer != "" && opts.SetContext {
				r = r.WithContext(context.SetReferer(r.Context(), referer))
			}

			next.ServeHTTP(w, r)
		})
	}
}
