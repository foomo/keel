package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	TokenAuthOptions struct {
		// TokenProvider function to retrieve the token
		TokenProvider TokenProvider
	}
	TokenAuthOption func(*TokenAuthOptions)
)

// GetDefaultTokenAuthOptions returns the default options
func GetDefaultTokenAuthOptions() TokenAuthOptions {
	return TokenAuthOptions{
		TokenProvider: HeaderTokenProvider(),
	}
}

// TokenAuthWithTokenProvider middleware option
func TokenAuthWithTokenProvider(v TokenProvider) TokenAuthOption {
	return func(o *TokenAuthOptions) {
		o.TokenProvider = v
	}
}

// TokenAuth middleware
func TokenAuth(token string, opts ...TokenAuthOption) Middleware {
	options := GetDefaultTokenAuthOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return TokenAuthWithOptions(token, options)
}

// TokenAuthWithOptions middleware
func TokenAuthWithOptions(token string, opts TokenAuthOptions) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if value, err := opts.TokenProvider(r); err != nil {
				httputils.UnauthorizedServerError(l, w, r, errors.Wrap(err, "failed to retrieve token"))
				return
			} else if value == "" {
				httputils.UnauthorizedServerError(l, w, r, errors.New("missing token"))
				return
			} else if subtle.ConstantTimeCompare([]byte(value), []byte(token)) != 1 {
				httputils.UnauthorizedServerError(l, w, r, errors.New("invalid token"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
