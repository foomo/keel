package cookie

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	keelhttp "github.com/foomo/keel/utils/net/http"
)

type DomainProvider func(r *http.Request) (string, error)

var ErrDomainNotAllowed = errors.New("domain not allowed")

type (
	DomainProviderOptions struct {
		Domains  []string
		Mappings map[string]string
	}
	DomainProviderOption func(options *DomainProviderOptions)
)

func GetDefaultDomainProviderOptions() DomainProviderOptions {
	return DomainProviderOptions{}
}

func DomainProviderWithDomains(v ...string) DomainProviderOption {
	return func(o *DomainProviderOptions) {
		o.Domains = v
	}
}

func DomainProviderWithMappings(v map[string]string) DomainProviderOption {
	return func(o *DomainProviderOptions) {
		o.Mappings = v
	}
}

func NewDomainProvider(opts ...DomainProviderOption) DomainProvider {
	options := GetDefaultDomainProviderOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return func(r *http.Request) (string, error) {
		domain := keelhttp.GetRequestDomain(r)

		if options.Domains == nil || len(options.Domains) == 0 {
			return domain, nil
		}

		if options.Mappings != nil {
			if value, ok := options.Mappings[domain]; ok {
				domain = value
			}
		}

		if options.Domains != nil && len(options.Domains) > 0 {
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
