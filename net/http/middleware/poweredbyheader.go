package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
)

type (
	PoweredByHeaderOptions struct {
		Header  string
		Message string
	}
	PoweredByHeaderOption func(*PoweredByHeaderOptions)
)

// GetDefaultPoweredByHeaderOptions returns the default options
func GetDefaultPoweredByHeaderOptions() PoweredByHeaderOptions {
	return PoweredByHeaderOptions{
		Header:  keelhttp.HeaderXPoweredBy,
		Message: "a lot of LOVE",
	}
}

// PoweredByHeader middleware
func PoweredByHeader(opts ...PoweredByHeaderOption) Middleware {
	options := GetDefaultPoweredByHeaderOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return PoweredByHeaderWithOptions(options)
}

// PoweredByHeaderWithHeader middleware option
func PoweredByHeaderWithHeader(v string) PoweredByHeaderOption {
	return func(o *PoweredByHeaderOptions) {
		o.Header = v
	}
}

// PoweredByHeaderWithMessage middleware option
func PoweredByHeaderWithMessage(v string) PoweredByHeaderOption {
	return func(o *PoweredByHeaderOptions) {
		o.Message = v
	}
}

// PoweredByHeaderWithOptions middleware
func PoweredByHeaderWithOptions(opts PoweredByHeaderOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("PowerdByHeader")
			}

			w.Header().Add(opts.Header, name)
			next.ServeHTTP(w, r)
		})
	}
}
