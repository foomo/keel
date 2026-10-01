package middleware

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	keelhttp "github.com/foomo/keel/net/http"
)

type (
	// HeaderTokenProviderOptions configures [HeaderTokenProvider].
	HeaderTokenProviderOptions struct {
		// Prefix is the required value prefix, stripped from the token.
		Prefix string
		// Header is the name of the request header holding the token.
		Header string
	}
	// HeaderTokenProviderOption configures [HeaderTokenProviderOptions].
	HeaderTokenProviderOption func(*HeaderTokenProviderOptions)
)

// GetDefaultHeaderTokenOptions returns the default options, which read a
// bearer token from the Authorization header.
func GetDefaultHeaderTokenOptions() HeaderTokenProviderOptions {
	return HeaderTokenProviderOptions{
		Prefix: keelhttp.HeaderValueAuthorizationPrefix,
		Header: keelhttp.HeaderAuthorization,
	}
}

// HeaderTokenProviderWithPrefix sets the required value prefix. Defaults to
// "Bearer ".
func HeaderTokenProviderWithPrefix(v string) HeaderTokenProviderOption {
	return func(o *HeaderTokenProviderOptions) {
		o.Prefix = v
	}
}

// HeaderTokenProviderWithHeader sets the request header holding the token.
// Defaults to Authorization.
func HeaderTokenProviderWithHeader(v string) HeaderTokenProviderOption {
	return func(o *HeaderTokenProviderOptions) {
		o.Header = v
	}
}

// HeaderTokenProvider returns a [TokenProvider] that reads the token from a
// request header and strips the configured prefix. A missing header yields an
// empty token and no error; a value without the prefix yields an error.
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
