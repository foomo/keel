package middleware

import (
	"net/http"
)

// RequestURIWhitelistSkipper returns a [Skipper] that skips all requests
// except those whose RequestURI, including the query, exactly matches one of
// uris.
func RequestURIWhitelistSkipper(uris ...string) Skipper {
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
