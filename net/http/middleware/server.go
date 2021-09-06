package middleware

import (
	"net/http"

	"go.uber.org/zap"

	httputils "github.com/foomo/keel/net/http"
)

type (
	ServerOptions struct {
		Header string
	}
	ServerOption func(*ServerOptions)
)

// GetDefaultServerOptions returns the default options
func GetDefaultServerOptions() ServerOptions {
	return ServerOptions{
		Header: httputils.HeaderServer,
	}
}

// Server middleware
func Server(name string, opts ...ServerOption) Middleware {
	options := GetDefaultServerOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return ServerWithOptions(name, options)
}

// ServerWithOptions middleware
func ServerWithOptions(name string, opts ServerOptions) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(opts.Header, name)
			next.ServeHTTP(w, r)
		})
	}
}
