package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	BasicAuthOptions struct {
		Realm string
	}
	BasicAuthOption func(*BasicAuthOptions)
)

// GetDefaultBasicAuthOptions returns the default options
func GetDefaultBasicAuthOptions() BasicAuthOptions {
	return BasicAuthOptions{
		Realm: "Restricted",
	}
}

// BasicAuthWithRealm middleware option
func BasicAuthWithRealm(v string) BasicAuthOption {
	return func(o *BasicAuthOptions) {
		o.Realm = v
	}
}

// BasicAuth middleware
func BasicAuth(username string, passwordHash []byte, opts ...BasicAuthOption) Middleware {
	options := GetDefaultBasicAuthOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return BasicAuthWithOptions(username, passwordHash, options)
}

// BasicAuthWithOptions middleware
func BasicAuthWithOptions(username string, passwordHash []byte, opts BasicAuthOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("BasicAuth")
			}

			// basic auth from request header
			u, p, ok := r.BasicAuth()
			if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%s", opts.Realm))
				httputils.UnauthorizedServerError(l, w, r, errors.New("missing basic auth credentials"))

				return
			}

			// Compare the username and password hash with the ones in the request
			userMatch := subtle.ConstantTimeCompare([]byte(u), []byte(username)) == 1

			errP := bcrypt.CompareHashAndPassword(passwordHash, []byte(p))
			if !userMatch || errP != nil {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%s", opts.Realm))
				httputils.UnauthorizedServerError(l, w, r, errors.New("invalid basic auth credentials"))

				return
			}

			// If required, Context could be updated to include authentication
			// related data so that it could be used in consequent steps.
			next.ServeHTTP(w, r)
		})
	}
}
