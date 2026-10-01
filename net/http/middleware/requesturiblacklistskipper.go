package middleware

import (
	"net/http"
)

// RequestURIBlacklistSkipper returns a [Skipper] that skips requests whose
// RequestURI, including the query, exactly matches one of uris.
func RequestURIBlacklistSkipper(uris ...string) Skipper {
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
