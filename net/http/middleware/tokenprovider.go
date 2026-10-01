package middleware

import (
	"net/http"
)

// TokenProvider extracts an authentication token from the request. It returns
// an empty token and a nil error when the request carries none.
type TokenProvider func(r *http.Request) (string, error)
