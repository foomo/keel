package middleware

import (
	"context"
	"net/http"

	jwt2 "github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/jwt"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	JWTOptions struct {
		TokenProvider       TokenProvider
		ClaimsProvider      JWTClaimsProvider
		MissingTokenHandler JWTMissingTokenHandler
		InvalidTokenHandler JWTInvalidTokenHandler
		ErrorHandler        JWTErrorHandler
	}
	JWTOption              func(*JWTOptions)
	JWTClaimsProvider      func() jwt2.Claims
	JWTErrorHandler        func(*zap.Logger, http.ResponseWriter, *http.Request, error) bool
	JWTMissingTokenHandler func(*zap.Logger, http.ResponseWriter, *http.Request) (jwt2.Claims, bool)
	JWTInvalidTokenHandler func(*zap.Logger, http.ResponseWriter, *http.Request, *jwt2.Token) bool
)

// DefaultJWTErrorHandler function
func DefaultJWTErrorHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) bool {
	httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed parse claims"))
	return false
}

// DefaultJWTMissingTokenHandler function
func DefaultJWTMissingTokenHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request) (jwt2.Claims, bool) {
	return nil, true
}

// RequiredJWTMissingTokenHandler function
func RequiredJWTMissingTokenHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request) (jwt2.Claims, bool) {
	httputils.BadRequestServerError(l, w, r, errors.New("missing jwt token"))
	return nil, false
}

// DefaultJWTInvalidTokenHandler function
func DefaultJWTInvalidTokenHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request, token *jwt2.Token) bool {
	httputils.BadRequestServerError(l, w, r, errors.New("invalid jwt token"))
	return false
}

// DefaultJWTClaimsProvider function
func DefaultJWTClaimsProvider() jwt2.Claims {
	return &jwt2.StandardClaims{}
}

// GetDefaultJWTOptions returns the default options
func GetDefaultJWTOptions() JWTOptions {
	return JWTOptions{
		TokenProvider:       HeaderTokenProvider(),
		ClaimsProvider:      DefaultJWTClaimsProvider,
		ErrorHandler:        DefaultJWTErrorHandler,
		InvalidTokenHandler: DefaultJWTInvalidTokenHandler,
		MissingTokenHandler: DefaultJWTMissingTokenHandler,
	}
}

// JWTWithTokenProvider middleware option
func JWTWithTokenProvider(v TokenProvider) JWTOption {
	return func(o *JWTOptions) {
		o.TokenProvider = v
	}
}

// JWTWithClaimsProvider middleware option
func JWTWithClaimsProvider(v JWTClaimsProvider) JWTOption {
	return func(o *JWTOptions) {
		o.ClaimsProvider = v
	}
}

// JWTWithInvalidTokenHandler middleware option
func JWTWithInvalidTokenHandler(v JWTInvalidTokenHandler) JWTOption {
	return func(o *JWTOptions) {
		o.InvalidTokenHandler = v
	}
}

// JWTWithMissingTokenHandler middleware option
func JWTWithMissingTokenHandler(v JWTMissingTokenHandler) JWTOption {
	return func(o *JWTOptions) {
		o.MissingTokenHandler = v
	}
}

// JWTWithErrorHandler middleware option
func JWTWithErrorHandler(v JWTErrorHandler) JWTOption {
	return func(o *JWTOptions) {
		o.ErrorHandler = v
	}
}

// JWT middleware
func JWT(jwt *jwt.JWT, contextKey interface{}, opts ...JWTOption) Middleware {
	options := GetDefaultJWTOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return JWTWithOptions(jwt, contextKey, options)
}

// JWTWithOptions middleware
func JWTWithOptions(jwt *jwt.JWT, contextKey interface{}, opts JWTOptions) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := opts.ClaimsProvider()
			if value := r.Context().Value(contextKey); value != nil {
				// TODO check if type matches the existing
				next.ServeHTTP(w, r)
			} else if value, err := opts.TokenProvider(r); err != nil {
				httputils.BadRequestServerError(l, w, r, errors.Wrap(err, "failed to retrieve token"))
			} else if value == "" {
				if claims, resume := opts.MissingTokenHandler(l, w, r); resume && claims != nil {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey, claims)))
				} else if resume {
					next.ServeHTTP(w, r)
				}
			} else if token, err := jwt.ParseWithClaims(value, claims); err != nil {
				// TODO check if type matches the existing
				if opts.ErrorHandler(l, w, r, err) {
					next.ServeHTTP(w, r)
				}
			} else if !token.Valid {
				if opts.InvalidTokenHandler(l, w, r, token) {
					next.ServeHTTP(w, r)
				}
			} else {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey, claims)))
			}
		})
	}
}
