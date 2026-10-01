package middleware

import (
	"net/http"
	"strings"

	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
)

type (
	// ServerHeaderOptions configures the [ServerHeader] middleware.
	ServerHeaderOptions struct {
		// Header is the name of the response header.
		Header string
		// Name is the header value. When empty, the service name is used.
		Name string
	}
	// ServerHeaderOption configures [ServerHeaderOptions].
	ServerHeaderOption func(*ServerHeaderOptions)
)

// GetDefaultServerHeaderOptions returns the default options using the Server
// header and the service name.
func GetDefaultServerHeaderOptions() ServerHeaderOptions {
	return ServerHeaderOptions{
		Header: keelhttp.HeaderServer,
	}
}

// ServerHeader returns a middleware that adds a header, Server by default,
// with the service name or the configured name to every response.
func ServerHeader(opts ...ServerHeaderOption) keelhttp.Middleware {
	options := GetDefaultServerHeaderOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return ServerHeaderWithOptions(options)
}

// ServerHeaderWithName sets the header value. Defaults to the service name.
func ServerHeaderWithName(v string) ServerHeaderOption {
	return func(o *ServerHeaderOptions) {
		o.Name = v
	}
}

// ServerHeaderWithHeader sets the response header name. Defaults to Server.
func ServerHeaderWithHeader(v string) ServerHeaderOption {
	return func(o *ServerHeaderOptions) {
		o.Header = v
	}
}

// ServerHeaderWithOptions is like [ServerHeader] but takes fully populated
// options.
func ServerHeaderWithOptions(opts ServerHeaderOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		if opts.Name != "" {
			name = opts.Name
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("ServerHeader",
					trace.WithAttributes(semconv.HTTPResponseHeader(strings.ToLower(opts.Header), name)),
				)
			}

			w.Header().Add(opts.Header, name)
			next.ServeHTTP(w, r)
		})
	}
}
