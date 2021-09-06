package middleware

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	keelhttp "github.com/foomo/keel/net/http"
)

type TokenProvider func(r *http.Request) (string, error)

type (
	HeaderTokenProviderOptions struct {
		Prefix string
		Header string
	}
	HeaderTokenProviderOption func(*HeaderTokenProviderOptions)
)

// GetDefaultHeaderTokenOptions returns the default options
func GetDefaultHeaderTokenOptions() HeaderTokenProviderOptions {
	return HeaderTokenProviderOptions{
		Prefix: keelhttp.HeaderValueAuthorizationPrefix,
		Header: keelhttp.HeaderAuthorization,
	}
}

// HeaderTokenProviderWithPrefix middleware option
func HeaderTokenProviderWithPrefix(v string) HeaderTokenProviderOption {
	return func(o *HeaderTokenProviderOptions) {
		o.Prefix = v
	}
}

// HeaderTokenProviderWithHeader middleware option
func HeaderTokenProviderWithHeader(v string) HeaderTokenProviderOption {
	return func(o *HeaderTokenProviderOptions) {
		o.Header = v
	}
}

func HeaderTokenProvider(opts ...HeaderTokenProviderOption) TokenProvider {
	options := GetDefaultHeaderTokenOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return func(r *http.Request) (string, error) {
		if value := r.Header.Get(options.Header); value == "" {
			return "", nil
		} else if !strings.HasPrefix(value, options.Prefix) {
			return "", errors.New("malformed bearer token")
		} else {
			return strings.TrimPrefix(value, options.Prefix), nil
		}
	}
}

type (
	CookieTokenProviderOptions struct{}
	CookieTokenProviderOption  func(*CookieTokenProviderOptions)
)

// GetDefaultCookieTokenOptions returns the default options
func GetDefaultCookieTokenOptions() CookieTokenProviderOptions {
	return CookieTokenProviderOptions{}
}

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
