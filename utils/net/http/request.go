package httputils

import (
	"net"
	"net/http"
	"strings"

	keelhttp "github.com/foomo/keel/net/http"
)

// GetRequestHost returns the request's host, preferring the X-Forwarded-Host
// header, then [http.Request.Host] for non-absolute request URLs, and the URL
// host otherwise. The result may include a port.
func GetRequestHost(r *http.Request) string {
	if value := r.Header.Get(keelhttp.HeaderXForwardedHost); value != "" {
		return value
	} else if !r.URL.IsAbs() {
		return r.Host
	} else {
		return r.URL.Host
	}
}

// GetRemoteAddr returns the client IP address of r. It prefers the
// X-Real-IP, True-Client-Ip and the first X-Forwarded-For entry, in that order,
// and falls back to the host part of [http.Request.RemoteAddr]. These headers
// are client-controlled unless set by a trusted proxy.
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

// GetRequestDomain returns the host returned by [GetRequestHost] with any
// port removed.
func GetRequestDomain(r *http.Request) string {
	domain := GetRequestHost(r)
	// right trim port
	if portIndex := strings.Index(domain, ":"); portIndex != -1 {
		domain = domain[:portIndex]
	}

	return domain
}
