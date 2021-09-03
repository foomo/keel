package middleware

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrDomainNotAllowed = errors.New("domain not allowed")
)

type DomainProvider func(r *http.Request) (string, error)

var DefaultDomainProvider = func(domains []string) DomainProvider {
	return func(r *http.Request) (string, error) {
		domain := getDomainFromHTTPRequest(r)
		if !isDomainAllowed(domain, domains) {
			return "", ErrDomainNotAllowed
		}
		return domain, nil
	}
}

var MappingDomainProvider = func(domains []string, mapping map[string]string) DomainProvider {
	return func(r *http.Request) (string, error) {
		domain := getDomainFromHTTPRequest(r)
		if value, ok := mapping[domain]; ok {
			domain = value
		}
		if !isDomainAllowed(domain, domains) {
			return "", errors.New("invalid domain: " + domain)
		}
		return domain, nil
	}
}

// getDomainFromHTTPRequest helper
func getDomainFromHTTPRequest(r *http.Request) string {
	var domain string
	if r.Header.Get("X-Forwarded-Host") != "" {
		domain = r.Header.Get("X-Forwarded-Host")
	} else if !r.URL.IsAbs() {
		domain = r.Host
	} else {
		domain = r.URL.Host
	}

	// right trim port
	portIndex := strings.Index(domain, ":")
	if portIndex != -1 {
		domain = domain[:portIndex]
	}

	return domain
}

// isDomainAllowed helper
func isDomainAllowed(domain string, domains []string) bool {
	if domains == nil || len(domains) == 0 {
		return true
	}
	for _, value := range domains {
		if domain == value || (strings.HasPrefix(value, "*.") && strings.HasSuffix(domain, value[2:])) {
			return true
		}
	}
	return false
}
