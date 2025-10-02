package httputils

import (
	"net"
	"net/http"
	"strings"

	keelhttp "github.com/foomo/keel/net/http"
)

// GetRequestHost returns the request's host
func GetRequestHost(r *http.Request) string {
	if value := r.Header.Get(keelhttp.HeaderXForwardedHost); value != "" {
		return value
	} else if !r.URL.IsAbs() {
		return r.Host
	} else {
		return r.URL.Host
	}
}

func GetRemoteAddr(r *http.Request) string {
	if value := r.Header.Get(keelhttp.HeaderXRealIP); value != "" {
		return value
	} else if value := r.Header.Get(keelhttp.HeaderTrueClientIP); value != "" {
		return value
	} else if value := r.Header.Get(keelhttp.HeaderXForwardedFor); value != "" {
		if i := strings.IndexAny(value, ", "); i > 0 {
			return value[:i]
		} else {
			return value
		}
	} else if value, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return value
	} else {
		return r.RemoteAddr
	}
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
