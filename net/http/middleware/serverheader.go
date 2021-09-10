package middleware

import (
	"net/http"

	"go.uber.org/zap"

	httputils "github.com/foomo/keel/net/http"
)

type (
	ServerHeaderOptions struct {
		Header string
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

// ServerHeaderWithHeader middleware option
func ServerHeaderWithHeader(v string) ServerHeaderOption {
	return func(o *ServerHeaderOptions) {
		o.Header = v
	}
}

// ServerHeaderWithOptions middleware
func ServerHeaderWithOptions(opts ServerHeaderOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(opts.Header, name)
			next.ServeHTTP(w, r)
		})
	}
}
