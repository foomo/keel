package middleware

import (
	"net/http"

	"go.uber.org/zap"

	httputils "github.com/foomo/keel/net/http"
)

type (
	ServerHeaderOptions struct {
		Header string
		Name   string
	}
	ServerHeaderOption func(*ServerHeaderOptions)
)

// GetDefaultServerHeaderOptions returns the default options
func GetDefaultServerHeaderOptions() ServerHeaderOptions {
	return ServerHeaderOptions{
		Header: httputils.HeaderServer,
	}
}

// ServerHeader middleware
func ServerHeader(opts ...ServerHeaderOption) Middleware {
	options := GetDefaultServerHeaderOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return ServerHeaderWithOptions(options)
}

// ServerHeaderWithName middleware option
func ServerHeaderWithName(v string) ServerHeaderOption {
	return func(o *ServerHeaderOptions) {
		o.Name = v
	}
}

// ServerHeaderWithHeader middleware option
func ServerHeaderWithHeader(v string) ServerHeaderOption {
	return func(o *ServerHeaderOptions) {
		o.Header = v
	}
}

// ServerHeaderWithOptions middleware
func ServerHeaderWithOptions(opts ServerHeaderOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		if opts.Name != "" {
			name = opts.Name
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(opts.Header, name)
			next.ServeHTTP(w, r)
		})
	}
}
