package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

func BearerAuth(bearerToken string) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(authHeader) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("malformed token")); err != nil {
					log.WithError(l, err).Error("failed to write http response")
				}
			} else {
				if authHeader[1] == bearerToken {
					next.ServeHTTP(w, r)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					if _, err := w.Write([]byte("Unauthorized")); err != nil {
						log.WithError(l, err).Error("failed to write http response")
					}
				}
			}
		})
	}
}

func BasicAuth(user, password string) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
			u, p, ok := rq.BasicAuth()
			if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
				unauthorised(rw)
				return
			}

			// This is a dummy check for credentials.
			if u != user || p != password {
				unauthorised(rw)
				return
			}

			// If required, Context could be updated to include authentication
			// related data so that it could be used in consequent steps.
			next.ServeHTTP(rw, rq)
		})
	}
}

func unauthorised(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
	w.WriteHeader(http.StatusUnauthorized)
}
