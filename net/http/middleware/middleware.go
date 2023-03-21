package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

// Middleware your way to handle requests
type Middleware func(l *zap.Logger, name string, next http.Handler) http.Handler

func Compose(l *zap.Logger, name string, handler http.Handler, middlewares ...Middleware) http.Handler {
	composed := func(l *zap.Logger, name string, next http.Handler) http.Handler {
		for _, middleware := range middlewares {
			next = middleware(l, name, next)
		}
		return next
	}
	return composed(l, name, handler)
}
