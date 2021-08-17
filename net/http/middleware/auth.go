package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/foomo/keel/log"
)

func BearerAuth(bearerToken string) Middleware {
	bearerPrefix := "Bearer "
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(bearerToken, bearerPrefix) {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("malformed token")); err != nil {
					log.WithError(l, err).Error("failed to write http response")
				}
				return
			}

			authHeader := strings.Replace(bearerToken, bearerPrefix, "", 1)
			if subtle.ConstantTimeCompare([]byte(authHeader), []byte(bearerToken)) == 1 {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte("Unauthorized")); err != nil {
				log.WithError(l, err).Error("failed to write http response")
			}
		})
	}
}

// BasicAuth hashes the password when called and returns a middleware.
// NOTE: The error handling only takes place on incomming http requests.
// Therefore (and because of security) it is adviced to hash the password
// beforehand and use BasicAuthBcryptHash.
func BasicAuth(user, password string) Middleware {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return func(l *zap.Logger, next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				l.Error("unable to create password hash", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			})
		}
	}

	return BasicAuthBcryptHash(user, string(hashedPassword))
}

// BasicAuthBcryptHash uses a plain text user name an a bcrypt salted hash of
// the password in order to authenticate the incomming http request.
func BasicAuthBcryptHash(user, hashedPassword string) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			u, p, ok := r.BasicAuth()
			if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
				unauthorised(w)
				return
			}

			// Compare the username and password hash with the ones in the request
			userMatch := (subtle.ConstantTimeCompare([]byte(u), []byte(user)) == 1)
			errP := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(p))
			if !userMatch || errP != nil {
				unauthorised(w)
				return
			}

			// If required, Context could be updated to include authentication
			// related data so that it could be used in consequent steps.
			next.ServeHTTP(w, r)
		})
	}
}

func unauthorised(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
	w.WriteHeader(http.StatusUnauthorized)
}
