package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func Skip(mw Middleware, skippers ...Skipper) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		wrapped := mw(l, name, next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, skipper := range skippers {
				if skipper(r) {
					next.ServeHTTP(w, r)
					return
				}
			}

			wrapped.ServeHTTP(w, r)
		})
	}
}
