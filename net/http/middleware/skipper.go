package middleware

import (
	"net/http"
)

type Skipper func(*http.Request) bool

// RequestURIWhitelistSkipper returns a HTTPMiddlewareConfig.Skipper which skips all but the given uris
func RequestURIWhitelistSkipper(uris ...string) Skipper {
	urisMap := make(map[string]bool, len(uris))
	for _, uri := range uris {
		urisMap[uri] = true
	}
	return func(r *http.Request) bool {
		if _, ok := urisMap[r.RequestURI]; ok {
			return true
		}
		return false
	}
}

// RequestURIBlacklistSkipper returns a HTTPMiddlewareConfig.Skipper which skips the given uris
func RequestURIBlacklistSkipper(uris ...string) Skipper {
	urisMap := make(map[string]bool, len(uris))
	for _, uri := range uris {
		urisMap[uri] = true
	}
	return func(r *http.Request) bool {
		if _, ok := urisMap[r.RequestURI]; ok {
			return false
		}
		return true
	}
}
