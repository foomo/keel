package http

import (
	"net/http"

	"go.uber.org/zap"
)

// Middleware wraps next with additional request handling. l is the logger and
// name the name of the service the handler belongs to.
type Middleware func(l *zap.Logger, name string, next http.Handler) http.Handler

// Compose wraps handler with the given middlewares and returns the result.
// Middlewares are applied in order, so the last one is the outermost and sees
// a request first.
func Compose(l *zap.Logger, name string, handler http.Handler, middlewares ...Middleware) http.Handler {
	composed := func(l *zap.Logger, name string, next http.Handler) http.Handler {
		for _, middleware := range middlewares {
			next = middleware(l, name, next)
		}

		return next
	}

	return composed(l, name, handler)
}
