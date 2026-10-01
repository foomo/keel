package middleware

import (
	"context"
	"net/http"

	keelhttp "github.com/foomo/keel/net/http"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/jwt"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	// JWTOptions configures the [JWT] middleware.
	JWTOptions struct {
		// SetContext reports whether a present token is parsed, validated and
		// its claims stored in the request context. When false, a present
		// token is neither parsed nor validated and the request is passed on
		// unchanged.
		SetContext bool
		// TokenProvider extracts the token from the request.
		TokenProvider TokenProvider
		// ClaimsProvider returns the claims value a token is parsed into.
		ClaimsProvider JWTClaimsProvider
		// ClaimsHandler is called with the claims of a valid token.
		ClaimsHandler JWTClaimsHandler
		// MissingTokenHandler is called when the request carries no token.
		MissingTokenHandler JWTMissingTokenHandler
		// InvalidTokenHandler is called when a token parses but is not valid.
		InvalidTokenHandler JWTInvalidTokenHandler
		// ErrorHandler is called when parsing or validating a token fails.
		ErrorHandler JWTErrorHandler
	}
	// JWTOption configures [JWTOptions].
	JWTOption func(*JWTOptions)
	// JWTClaimsProvider returns a new, empty claims value to parse a token
	// into. It is called once per request.
	JWTClaimsProvider func() gojwt.Claims
	// JWTClaimsHandler is called with the claims of a valid token and reports
	// whether the request continues, with the claims in its context.
	JWTClaimsHandler func(*zap.Logger, http.ResponseWriter, *http.Request, gojwt.Claims) bool
	// JWTErrorHandler is called when parsing a token fails and reports whether
	// the request continues without claims.
	JWTErrorHandler func(*zap.Logger, http.ResponseWriter, *http.Request, error) bool
	// JWTMissingTokenHandler is called when the request carries no token. It
	// returns optional claims to store in the context and reports whether the
	// request continues.
	JWTMissingTokenHandler func(*zap.Logger, http.ResponseWriter, *http.Request) (gojwt.Claims, bool)
	// JWTInvalidTokenHandler is called when a token parses but is not valid
	// and reports whether the request continues without claims.
	JWTInvalidTokenHandler func(*zap.Logger, http.ResponseWriter, *http.Request, *gojwt.Token) bool
)

// DefaultJWTErrorHandler responds with 500 Internal Server Error and stops
// the request.
func DefaultJWTErrorHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) bool {
	httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed parse claims"))
	return false
}

// DefaultJWTMissingTokenHandler lets the request continue without claims.
func DefaultJWTMissingTokenHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request) (gojwt.Claims, bool) {
	return nil, true
}

// RequiredJWTMissingTokenHandler responds with 400 Bad Request and stops the
// request. Use it to make a token mandatory.
func RequiredJWTMissingTokenHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request) (gojwt.Claims, bool) {
	httputils.BadRequestServerError(l, w, r, errors.New("missing jwt token"))
	return nil, false
}

// DefaultJWTInvalidTokenHandler responds with 400 Bad Request and stops the
// request.
func DefaultJWTInvalidTokenHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request, token *gojwt.Token) bool {
	httputils.BadRequestServerError(l, w, r, errors.New("invalid jwt token"))
	return false
}

// DefaultJWTClaimsProvider returns an empty [gojwt.RegisteredClaims].
func DefaultJWTClaimsProvider() gojwt.Claims {
	return &gojwt.RegisteredClaims{}
}

// DefaultJWTClaimsHandler accepts all claims.
func DefaultJWTClaimsHandler(l *zap.Logger, w http.ResponseWriter, r *http.Request, claims gojwt.Claims) bool {
	return true
}

// GetDefaultJWTOptions returns the default options: SetContext enabled, a
// [HeaderTokenProvider] and the Default* handlers of this package.
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

// JWTWithTokenProvider sets [JWTOptions.TokenProvider].
func JWTWithTokenProvider(v TokenProvider) JWTOption {
	return func(o *JWTOptions) {
		o.TokenProvider = v
	}
}

// JWTWithClaimsProvider sets [JWTOptions.ClaimsProvider].
func JWTWithClaimsProvider(v JWTClaimsProvider) JWTOption {
	return func(o *JWTOptions) {
		o.ClaimsProvider = v
	}
}

// JWTWithClaimsHandler sets [JWTOptions.ClaimsHandler].
func JWTWithClaimsHandler(v JWTClaimsHandler) JWTOption {
	return func(o *JWTOptions) {
		o.ClaimsHandler = v
	}
}

// JWTWithInvalidTokenHandler sets [JWTOptions.InvalidTokenHandler].
func JWTWithInvalidTokenHandler(v JWTInvalidTokenHandler) JWTOption {
	return func(o *JWTOptions) {
		o.InvalidTokenHandler = v
	}
}

// JWTWithMissingTokenHandler sets [JWTOptions.MissingTokenHandler].
func JWTWithMissingTokenHandler(v JWTMissingTokenHandler) JWTOption {
	return func(o *JWTOptions) {
		o.MissingTokenHandler = v
	}
}

// JWTWithErrorHandler sets [JWTOptions.ErrorHandler].
func JWTWithErrorHandler(v JWTErrorHandler) JWTOption {
	return func(o *JWTOptions) {
		o.ErrorHandler = v
	}
}

// JWTWithSetContext sets [JWTOptions.SetContext]. Defaults to true.
func JWTWithSetContext(v bool) JWTOption {
	return func(o *JWTOptions) {
		o.SetContext = v
	}
}

// JWT returns a middleware that parses and validates the request's JWT with v
// and stores the claims in the request context under contextKey. Requests
// whose context already holds a value for contextKey are passed on unchanged.
// A failing [TokenProvider] is answered with 400 Bad Request; all other
// outcomes are delegated to the handlers in [JWTOptions].
func JWT(v *jwt.JWT, contextKey any, opts ...JWTOption) keelhttp.Middleware {
	options := GetDefaultJWTOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return JWTWithOptions(v, contextKey, options)
}

// JWTWithOptions is like [JWT] but takes fully populated options.
func JWTWithOptions(v *jwt.JWT, contextKey any, opts JWTOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("JWT")
			}

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
