package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

// Middleware your way to handle requests
type (
	Middleware func(*zap.Logger, http.Handler) http.Handler
)

func Compose(l *zap.Logger, handler http.Handler, middlewares ...Middleware) http.Handler {
	composed := func(l *zap.Logger, next http.Handler) http.Handler {
		for _, middleware := range middlewares {
			next = middleware(l, next)
		}
		return next
	}
	return composed(l, handler)
}
