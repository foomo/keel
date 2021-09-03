package middleware

import (
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	httputils "github.com/foomo/keel/net/http"
)

type (
	SessionIDConfig struct {
		SetCookie            bool // only create a session cookie if enabled
		CookieName           string
		CookieSecure         bool
		CookieHttpOnly       bool
		CookiePath           string
		CookieDomain         string
		CookieDomains        []string
		CookieDomainProvider DomainProvider
		Generator            SessionIDGenerator
	}
	SessionIDOption func(*SessionIDConfig) error
)

var DefaultSessionIDConfig = SessionIDConfig{
	SetCookie:      false,
	CookieName:     "keel-session",
	CookieSecure:   true,
	CookieHttpOnly: true,
	CookiePath:     "/",
}

func SessionIDWithSetCookie(v bool) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.SetCookie = v
	}
}

func SessionIDWithCookieName(v string) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookieName = v
	}
}

func SessionIDWithCookieSecure(v bool) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookieSecure = v
	}
}

func SessionIDWithCookieHttpOnly(v bool) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookieHttpOnly = v
	}
}

func SessionIDWithCookiePath(v string) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookiePath = v
	}
}

func SessionIDWithCookieDomain(v string) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookieDomain = v
	}
}

func SessionIDWithCookieDomains(v []string) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookieDomains = v
	}
}

func SessionIDWithCookieDomainProvider(v DomainProvider) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.CookieDomainProvider = v
	}
}

func SessionIDWithGenerator(v SessionIDGenerator) SessionIDOption {
	return func(c *SessionIDConfig) error {
		c.Generator = v
	}
}

func SessionID(opts ...SessionIDOption) Middleware {
	config = DefaultSessionIDConfig
	for _, opt := range opts {
		if opt != nil {
			if err := opt(&opts); err != nil {
				return nil, err
			}
		}
	}
	if config.Generator == nil {
		config.Generator = DefaultSessionIDGenerator
	}
	if config.DomainProvider == nil {
		config.DomainProvider = DefaultDomainProvider(config.CookieDomains)
	}

	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if id := r.Header.Get(httputils.HeaderXSessionID); id == "" {
				if cookie, err := r.Cookie(config.CookieName); errors.Is(err, http.ErrNoCookie) {
					if config.SetCookie {

						domain, err := config.DomainProvider(r)
						if err != nil {
							httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to resolve domain"))
							return
						}

						id = config.Generator()

						http.SetCookie(w, &http.Cookie{
							Name:     config.CookieName,
							Value:    id,
							Path:     config.CookiePath,
							HttpOnly: config.CookieHttpOnly,
							Secure:   config.CookieSecure,
							Domain:   domain,
						})
					}
					r.Header.Set(httputils.HeaderXSessionID, id)
				} else if err != nil {
					httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to read cookie"))
					return
				} else {
					r.Header.Set(httputils.HeaderXSessionID, cookie.Value)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
