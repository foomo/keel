package cookie

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type (
	Cookie struct {
		Name string
		// Path the path of the created cookie
		Path string
		// Secure the secure flag of the created cookie
		Secure bool
		// MaxAge the max age flag of the created cookie
		MaxAge int
		// Expires the expires flag of the created cookie
		Expires time.Duration
		// Secure the secure flag of the created cookie
		SameSite http.SameSite
		// HTTPOnly the http only of the created cookie
		HTTPOnly bool
		// TimeProvider function to retrieve the now time for the expires flag of the created cookie
		TimeProvider TimeProvider
		// DomainProvider function to retrieve the domain flag of the created cookie
		DomainProvider DomainProvider
	}
	Option func(options *Cookie)
)

// WithSecure middleware option
func WithSecure(v bool) Option {
	return func(o *Cookie) {
		o.Secure = v
	}
}

// WithHTTPOnly middleware option
func WithHTTPOnly(v bool) Option {
	return func(o *Cookie) {
		o.HTTPOnly = v
	}
}

// WithMaxAge middleware option
func WithMaxAge(v int) Option {
	return func(o *Cookie) {
		o.MaxAge = v
	}
}

// WithExpires middleware option
func WithExpires(v time.Duration) Option {
	return func(o *Cookie) {
		o.Expires = v
	}
}

// WithPath middleware option
func WithPath(v string) Option {
	return func(o *Cookie) {
		o.Path = v
	}
}

// WithSameSite middleware option
func WithSameSite(v http.SameSite) Option {
	return func(o *Cookie) {
		o.SameSite = v
	}
}

// WithTimeProvider middleware option
func WithTimeProvider(v TimeProvider) Option {
	return func(o *Cookie) {
		o.TimeProvider = v
	}
}

// WithDomainProvider middleware option
func WithDomainProvider(v DomainProvider) Option {
	return func(o *Cookie) {
		o.DomainProvider = v
	}
}

// New return a new provider
func New(name string, opts ...Option) Cookie {
	inst := Cookie{
		Name:     name,
		Path:     "/",
		Secure:   true,
		HTTPOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&inst)
		}
	}
	if inst.DomainProvider == nil {
		inst.DomainProvider = NewDomainProvider()
	}
	return inst
}

func (c Cookie) Delete(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(c.Name); errors.Is(err, http.ErrNoCookie) {
		return nil
	} else if err != nil {
		return err
	} else {
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}
	return nil
}

func (c Cookie) Get(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(c.Name)
}

func (c Cookie) Set(w http.ResponseWriter, r *http.Request, value string, opts ...Option) error {
	domain, err := c.DomainProvider(r)
	if err != nil {
		return err
	}
	options := c
	for _, opt := range opts {
		opt(&options)
	}
	cookie := &http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     options.Path,
		Domain:   domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}
	if options.Expires.Nanoseconds() > 0 {
		cookie.Expires = options.TimeProvider().Add(options.Expires)
	}
	http.SetCookie(w, cookie)
	return nil
}
