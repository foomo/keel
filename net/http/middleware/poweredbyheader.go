package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
)

type (
	// PoweredByHeaderOptions configures the [PoweredByHeader] middleware.
	PoweredByHeaderOptions struct {
		// Header is the name of the response header.
		Header string
		// Message is not used; the header value is the service name.
		Message string
	}
	// PoweredByHeaderOption configures [PoweredByHeaderOptions].
	PoweredByHeaderOption func(*PoweredByHeaderOptions)
)

// GetDefaultPoweredByHeaderOptions returns the default options using the
// X-Powered-By header.
func GetDefaultPoweredByHeaderOptions() PoweredByHeaderOptions {
	return PoweredByHeaderOptions{
		Header:  keelhttp.HeaderXPoweredBy,
		Message: "a lot of LOVE",
	}
}

// PoweredByHeader returns a middleware that adds a header, X-Powered-By by
// default, with the service name to every response.
func PoweredByHeader(opts ...PoweredByHeaderOption) keelhttp.Middleware {
	options := GetDefaultPoweredByHeaderOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return PoweredByHeaderWithOptions(options)
}

// PoweredByHeaderWithHeader sets the response header name. Defaults to
// X-Powered-By.
func PoweredByHeaderWithHeader(v string) PoweredByHeaderOption {
	return func(o *PoweredByHeaderOptions) {
		o.Header = v
	}
}

// PoweredByHeaderWithMessage sets [PoweredByHeaderOptions.Message].
func PoweredByHeaderWithMessage(v string) PoweredByHeaderOption {
	return func(o *PoweredByHeaderOptions) {
		o.Message = v
	}
}

// PoweredByHeaderWithOptions is like [PoweredByHeader] but takes fully
// populated options.
func PoweredByHeaderWithOptions(opts PoweredByHeaderOptions) keelhttp.Middleware {
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
