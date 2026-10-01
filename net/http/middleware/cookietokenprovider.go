package middleware

import (
	"net/http"

	"github.com/pkg/errors"
)

type (
	// CookieTokenProviderOptions configures [CookieTokenProvider]. It has no
	// fields and exists for API symmetry with the other token providers.
	CookieTokenProviderOptions struct{}
	// CookieTokenProviderOption configures [CookieTokenProviderOptions].
	CookieTokenProviderOption func(*CookieTokenProviderOptions)
)

// GetDefaultCookieTokenOptions returns the default options.
func GetDefaultCookieTokenOptions() CookieTokenProviderOptions {
	return CookieTokenProviderOptions{}
}

// CookieTokenProvider returns a [TokenProvider] that reads the token from the
// cookie named cookieName. A missing cookie yields an empty token and no error.
func CookieTokenProvider(cookieName string, opts ...CookieTokenProviderOption) TokenProvider {
	options := GetDefaultCookieTokenOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return func(r *http.Request) (string, error) {
		if cookie, err := r.Cookie(cookieName); errors.Is(err, http.ErrNoCookie) {
			return "", nil
		} else if err != nil {
			return "", errors.New("failed to retrieve cookie")
		} else {
			return cookie.Value, nil
		}
	}
}
