package middleware

import (
	"context"
	"fmt"
	"net/http"

	jwt2 "github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/jwt"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	JWTOptions struct {
		SetContext          bool
		TokenProvider       TokenProvider
		ClaimsProvider      JWTClaimsProvider
		ClaimsHandler       JWTClaimsHandler
		MissingTokenHandler JWTMissingTokenHandler
		InvalidTokenHandler JWTInvalidTokenHandler
		ErrorHandler        JWTErrorHandler
	}
	JWTOption              func(*JWTOptions)
	JWTClaimsProvider      func() jwt2.Claims
	JWTClaimsHandler       func(*zap.Logger, http.ResponseWriter, *http.Request, jwt2.Claims) bool
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

// DefaultJWTClaimsHandler function
func DefaultJWTClaimsHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request, claims jwt2.Claims) bool {
	return true
}

// GetDefaultJWTOptions returns the default options
func GetDefaultJWTOptions() JWTOptions {
	return JWTOptions{
		SetContext:          true,
		TokenProvider:       HeaderTokenProvider(),
		ClaimsProvider:      DefaultJWTClaimsProvider,
		ClaimsHandler:       DefaultJWTClaimsHandler,
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

// JWTWithClaimsHandler middleware option
func JWTWithClaimsHandler(v JWTClaimsHandler) JWTOption {
	return func(o *JWTOptions) {
		o.ClaimsHandler = v
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

func JWTWithSetContext(v bool) JWTOption {
	return func(o *JWTOptions) {
		o.SetContext = v
	}
}

// JWT middleware
func JWT(v *jwt.JWT, contextKey interface{}, opts ...JWTOption) Middleware {
	options := GetDefaultJWTOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return JWTWithOptions(v, contextKey, options)
}

// JWTWithOptions middleware
func JWTWithOptions(v *jwt.JWT, contextKey interface{}, opts JWTOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			span.AddEvent("JWT")
			fmt.Println("jwt:", span.SpanContext().SpanID())

			claims := opts.ClaimsProvider()

			// check existing claims from context
			if value := r.Context().Value(contextKey); value != nil {
				next.ServeHTTP(w, r)
				return
			}

			// retrieve token from provider
			token, err := opts.TokenProvider(r)
			if err != nil {
				httputils.BadRequestServerError(l, w, r, errors.Wrap(err, "failed to retrieve token"))
				return
			}

			// handle missing token
			if token == "" {
				if claims, resume := opts.MissingTokenHandler(l, w, r); claims != nil && resume && opts.SetContext {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey, claims)))
					return
				} else if resume {
					next.ServeHTTP(w, r)
					return
				} else {
					return
				}
			}

			// don't validate if not required
			if !opts.SetContext {
				next.ServeHTTP(w, r)
				return
			}

			// handle existing token
			jwtToken, err := v.ParseWithClaims(token, claims)
			if err != nil {
				if resume := opts.ErrorHandler(l, w, r, err); resume {
					next.ServeHTTP(w, r)
					return
				} else {
					return
				}
			} else if !jwtToken.Valid {
				if resume := opts.InvalidTokenHandler(l, w, r, jwtToken); resume {
					next.ServeHTTP(w, r)
					return
				} else {
					return
				}
			} else if resume := opts.ClaimsHandler(l, w, r, claims); !resume {
				return
			} else {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey, claims)))
				return
			}
		})
	}
}
