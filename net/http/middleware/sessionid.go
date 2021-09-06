package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/cookie"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	contextKey       string
	SessionIDOptions struct {
		// Header to look up the session id
		Header string
		// Cookie how to set the cookie
		Cookie cookie.Cookie
		// Generator for the session ids
		Generator SessionIDGenerator
		// SetCookie if true it will create a cookie if not exists
		SetCookie bool
		// SetHeader if true it will set add a request header
		SetHeader bool
		// SetContext if true it will set the context key
		SetContext bool
	}
	SessionIDOption    func(*SessionIDOptions)
	SessionIDGenerator func() string
)

const (
	ContextKeySessionID contextKey = "sessionId"

	DefaultSessionIDCookieName = "keel-session"
)

// DefaultSessionIDGenerator function
func DefaultSessionIDGenerator() string {
	return uuid.New().String()
}

// GetDefaultSessionIDOptions returns the default options
func GetDefaultSessionIDOptions() SessionIDOptions {
	return SessionIDOptions{
		Header:     keelhttp.HeaderXSessionID,
		Cookie:     cookie.New(DefaultSessionIDCookieName),
		Generator:  DefaultSessionIDGenerator,
		SetCookie:  false,
		SetHeader:  true,
		SetContext: true,
	}
}

func SessionIDWithHeader(v string) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Header = v
	}
}

// SessionIDWithSetCookie middleware option
func SessionIDWithSetCookie(v bool) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.SetCookie = v
	}
}

// SessionIDWithSetHeader middleware option
func SessionIDWithSetHeader(v bool) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.SetHeader = v
	}
}

// SessionIDWithSetContext middleware option
func SessionIDWithSetContext(v bool) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.SetContext = v
	}
}

// SessionIDWithCookie middleware option
func SessionIDWithCookie(v cookie.Cookie) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Cookie = v
	}
}

// SessionIDWithGenerator middleware option
func SessionIDWithGenerator(v SessionIDGenerator) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Generator = v
	}
}

// SessionID middleware
func SessionID(opts ...SessionIDOption) Middleware {
	options := GetDefaultSessionIDOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return SessionIDWithOptions(options)
}

// SessionIDWithOptions middleware
func SessionIDWithOptions(opts SessionIDOptions) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionID string
			if c, err := opts.Cookie.Get(r); errors.Is(err, http.ErrNoCookie) && !opts.SetCookie {
				// do nothing
			} else if errors.Is(err, http.ErrNoCookie) && opts.SetCookie {
				sessionID = opts.Generator()
				if err := opts.Cookie.Set(w, r, sessionID); err != nil {
					httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to set session id cookie"))
					return
				}
			} else if err != nil {
				httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to read session id cookie"))
				return
			} else {
				sessionID = c.Value
			}
			if sessionID != "" && opts.SetHeader {
				r.Header.Set(opts.Header, sessionID)
			}
			if sessionID != "" && opts.SetContext {
				r = r.WithContext(context.WithValue(r.Context(), ContextKeySessionID, sessionID))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SessionIDFromContext helper
func SessionIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(ContextKeySessionID).(string); ok {
		return value
	}
	return ""
}
