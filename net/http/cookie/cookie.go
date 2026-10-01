package cookie

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type (
	// Cookie describes a named HTTP cookie and the attributes used when it is
	// written by [Cookie.Set]. Create one with [New].
	Cookie struct {
		// Name the name of the cookie
		Name string
		// Path the path of the created cookie
		Path string
		// Secure the secure flag of the created cookie
		Secure bool
		// MaxAge the max age flag of the created cookie
		MaxAge int
		// Expires the expires flag of the created cookie
		Expires time.Duration
		// SameSite the same site flag of the created cookie
		SameSite http.SameSite
		// HTTPOnly the http only of the created cookie
		HTTPOnly bool
		// TimeProvider function to retrieve the now time for the expires flag of the created cookie
		TimeProvider TimeProvider
		// DomainProvider function to retrieve the domain flag of the created cookie
		DomainProvider DomainProvider
	}
	// Option configures a [Cookie] in [New] or overrides its attributes in
	// [Cookie.Set].
	Option func(options *Cookie)
)

// WithSecure sets the Secure attribute. Defaults to true.
func WithSecure(v bool) Option {
	return func(o *Cookie) {
		o.Secure = v
	}
}

// WithHTTPOnly sets the HttpOnly attribute. Defaults to true.
func WithHTTPOnly(v bool) Option {
	return func(o *Cookie) {
		o.HTTPOnly = v
	}
}

// WithMaxAge sets the Max-Age attribute in seconds; see [http.Cookie.MaxAge]
// for the meaning of zero and negative values. Defaults to 0.
func WithMaxAge(v int) Option {
	return func(o *Cookie) {
		o.MaxAge = v
	}
}

// WithExpires sets the lifetime used to compute the Expires attribute
// relative to the [TimeProvider]. Non-positive values omit Expires, which is
// the default.
func WithExpires(v time.Duration) Option {
	return func(o *Cookie) {
		o.Expires = v
	}
}

// WithPath sets the Path attribute. Defaults to "/".
func WithPath(v string) Option {
	return func(o *Cookie) {
		o.Path = v
	}
}

// WithSameSite sets the SameSite attribute. Defaults to
// [http.SameSiteDefaultMode].
func WithSameSite(v http.SameSite) Option {
	return func(o *Cookie) {
		o.SameSite = v
	}
}

// WithTimeProvider sets the clock used to compute the Expires attribute.
// Defaults to [NewTimeProvider] without options.
func WithTimeProvider(v TimeProvider) Option {
	return func(o *Cookie) {
		o.TimeProvider = v
	}
}

// WithDomainProvider sets the function resolving the Domain attribute from the
// request. Defaults to [NewDomainProvider] without options.
func WithDomainProvider(v DomainProvider) Option {
	return func(o *Cookie) {
		o.DomainProvider = v
	}
}

// New returns a [Cookie] with the given name, configured by opts. Without
// options the cookie uses path "/", is Secure and HttpOnly, and uses
// [http.SameSiteDefaultMode]. Nil options are ignored.
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

	if inst.TimeProvider == nil {
		inst.TimeProvider = NewTimeProvider()
	}

	return inst
}

// Delete expires the cookie on the client by writing it with an empty value
// and a negative Max-Age. It does nothing if r does not carry the cookie and
// returns any error from the domain provider.
func (c Cookie) Delete(w http.ResponseWriter, r *http.Request) error {
	if _, err := r.Cookie(c.Name); errors.Is(err, http.ErrNoCookie) {
		return nil
	} else if err != nil {
		return err
	} else if _, err := c.Set(w, r, "", WithMaxAge(-1)); err != nil {
		return err
	}

	return nil
}

// Get returns the cookie from r, or [http.ErrNoCookie] if it is not present.
func (c Cookie) Get(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(c.Name)
}

// Set writes the cookie with the given value to w and returns it. opts
// override the cookie's attributes for this call only; the name and domain
// provider of c are always used. It returns an error if the domain provider
// rejects the request's domain.
func (c Cookie) Set(w http.ResponseWriter, r *http.Request, value string, opts ...Option) (*http.Cookie, error) {
	domain, err := c.DomainProvider(r)
	if err != nil {
		return nil, err
	}

	options := c
	for _, opt := range opts {
		opt(&options)
	}

	cookie := &http.Cookie{ //nolint:gosec // G124: provided by options
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

	return cookie, nil
}
