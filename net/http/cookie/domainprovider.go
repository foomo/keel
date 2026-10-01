package cookie

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	keelhttp "github.com/foomo/keel/utils/net/http"
)

// DomainProvider returns the cookie domain to use for the request r.
type DomainProvider func(r *http.Request) (string, error)

// ErrDomainNotAllowed is returned by a [DomainProvider] created by
// [NewDomainProvider] when the request domain matches none of the allowed
// domains.
var ErrDomainNotAllowed = errors.New("domain not allowed")

type (
	// DomainProviderOptions configures [NewDomainProvider].
	DomainProviderOptions struct {
		// Domains lists the allowed domains; a "*." prefix also matches any
		// subdomain. An empty list allows every domain.
		Domains []string
		// Mappings maps a request domain to the cookie domain to use instead.
		// It is only consulted when Domains is not empty.
		Mappings map[string]string
	}
	// DomainProviderOption configures [DomainProviderOptions].
	DomainProviderOption func(options *DomainProviderOptions)
)

// GetDefaultDomainProviderOptions returns the default options, which allow
// every domain.
func GetDefaultDomainProviderOptions() DomainProviderOptions {
	return DomainProviderOptions{}
}

// DomainProviderWithDomains sets the allowed domains.
func DomainProviderWithDomains(v ...string) DomainProviderOption {
	return func(o *DomainProviderOptions) {
		o.Domains = v
	}
}

// DomainProviderWithMappings sets the domain mappings.
func DomainProviderWithMappings(v map[string]string) DomainProviderOption {
	return func(o *DomainProviderOptions) {
		o.Mappings = v
	}
}

// NewDomainProvider returns a [DomainProvider] that derives the domain from the
// request using [keelhttp.GetRequestDomain]. If allowed domains are
// configured, the domain is first replaced by its mapping, if any, and
// [ErrDomainNotAllowed] is returned unless it matches one of them.
func NewDomainProvider(opts ...DomainProviderOption) DomainProvider {
	options := GetDefaultDomainProviderOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return func(r *http.Request) (string, error) {
		domain := keelhttp.GetRequestDomain(r)

		if len(options.Domains) == 0 {
			return domain, nil
		}

		if options.Mappings != nil {
			if value, ok := options.Mappings[domain]; ok {
				domain = value
			}
		}

		if len(options.Domains) > 0 {
			for _, value := range options.Domains {
				// foo.com = foo.com
				// foo.com = *.foo.com
				// example.foo.com = *.foo.com
				if domain == strings.TrimPrefix(value, "*.") || (strings.HasPrefix(value, "*.") && strings.HasSuffix(domain, value[1:])) {
					return domain, nil
				}
			}
		}

		return "", ErrDomainNotAllowed
	}
}
