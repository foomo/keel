package middleware

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	keelhttp "github.com/foomo/keel/net/http"
)

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
