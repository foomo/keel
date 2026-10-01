package middleware

import (
	"net/http"
)

// Skipper reports whether a middleware should be skipped for the request.
type Skipper func(*http.Request) bool
