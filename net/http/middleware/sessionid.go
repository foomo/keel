package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/cookie"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	// SessionIDOptions configures the [SessionID] middleware.
	SessionIDOptions struct {
		// Header is the request header the session ID is read from and, with
		// SetHeader, written to.
		Header string
		// Cookie is the cookie the session ID is read from and, with SetCookie,
		// written to.
		Cookie cookie.Cookie
		// Generator creates a session ID when SetCookie is enabled and the
		// request carries none.
		Generator SessionIDGenerator
		// SetCookie reports whether a session ID is generated and set as cookie
		// when the request carries neither header nor cookie.
		SetCookie bool
		// SetHeader reports whether the session ID is set as request header.
		SetHeader bool
		// SetContext reports whether the session ID is stored in the request
		// context.
		SetContext bool
	}
	// SessionIDOption configures [SessionIDOptions].
	SessionIDOption func(*SessionIDOptions)
	// SessionIDGenerator returns a new session ID.
	SessionIDGenerator func() string
)

const (
	// DefaultSessionIDCookieName is the default name of the session ID cookie.
	DefaultSessionIDCookieName = "keel-session"
)

// DefaultSessionIDGenerator returns a random UUID string.
func DefaultSessionIDGenerator() string {
	return uuid.New().String()
}

// GetDefaultSessionIDOptions returns the default options: the X-Session-ID header, a
// cookie named "keel-session", UUID generation, SetHeader and SetContext enabled
// and SetCookie disabled.
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

// SessionIDWithHeader sets [SessionIDOptions.Header]. Defaults to X-Session-ID.
func SessionIDWithHeader(v string) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Header = v
	}
}

// SessionIDWithSetCookie sets [SessionIDOptions.SetCookie]. Defaults to false.
func SessionIDWithSetCookie(v bool) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.SetCookie = v
	}
}

// SessionIDWithSetHeader sets [SessionIDOptions.SetHeader]. Defaults to true.
func SessionIDWithSetHeader(v bool) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.SetHeader = v
	}
}

// SessionIDWithSetContext sets [SessionIDOptions.SetContext]. Defaults to true.
func SessionIDWithSetContext(v bool) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.SetContext = v
	}
}

// SessionIDWithCookie sets [SessionIDOptions.Cookie].
func SessionIDWithCookie(v cookie.Cookie) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Cookie = v
	}
}

// SessionIDWithGenerator sets [SessionIDOptions.Generator]. Defaults to
// [DefaultSessionIDGenerator].
func SessionIDWithGenerator(v SessionIDGenerator) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Generator = v
	}
}

// SessionID returns a middleware that resolves the session ID from the request
// header or, failing that, the cookie. With SetCookie, a missing cookie is
// generated and set on both response and request. Cookie errors are answered
// with 500 Internal Server Error.
func SessionID(opts ...SessionIDOption) keelhttp.Middleware {
	options := GetDefaultSessionIDOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return SessionIDWithOptions(options)
}

// SessionIDWithOptions is like [SessionID] but takes fully populated options.
func SessionIDWithOptions(opts SessionIDOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("SessionID")
			}

			var sessionID string
			if value := r.Header.Get(opts.Header); value != "" {
				sessionID = value
			} else if c, err := opts.Cookie.Get(r); errors.Is(err, http.ErrNoCookie) && !opts.SetCookie {
				// do nothing
			} else if errors.Is(err, http.ErrNoCookie) && opts.SetCookie {
				sessionID = opts.Generator()
				if c, err := opts.Cookie.Set(w, r, sessionID); err != nil {
					httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to set session id cookie"))
					return
				} else {
					r.AddCookie(c)
				}
			} else if err != nil {
				httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to read session id cookie"))
				return
			} else {
				sessionID = c.Value
			}

			if span.IsRecording() && sessionID != "" {
				span.SetAttributes(semconv.SessionID(sessionID))
			}

			if sessionID != "" && opts.SetHeader {
				r.Header.Set(opts.Header, sessionID)
			}

			if sessionID != "" && opts.SetContext {
				r = r.WithContext(keelhttpcontext.SetSessionID(r.Context(), sessionID))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SessionIDFromContext returns the session ID stored in ctx, or "" if there is none.
func SessionIDFromContext(ctx context.Context) string {
	if value, ok := keelhttpcontext.GetSessionID(ctx); ok {
		return value
	}

	return ""
}
