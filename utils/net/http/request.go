package httputils

import (
	"net/http"
	"strings"
)

// GetRequestHost returns the request's host
func GetRequestHost(r *http.Request) string {
	var host string
	switch {
	case r.Header.Get("X-Forwarded-Host") != "":
		host = r.Header.Get("X-Forwarded-Host")
	case !r.URL.IsAbs():
		host = r.Host
	default:
		host = r.URL.Host
	}
	return host
}

// GetRequestDomain returns the request's domain
func GetRequestDomain(r *http.Request) string {
	domain := GetRequestHost(r)
	// right trim port
	if portIndex := strings.Index(domain, ":"); portIndex != -1 {
		domain = domain[:portIndex]
	}
	return domain
}
