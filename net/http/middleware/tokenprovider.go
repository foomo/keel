package middleware

import (
	"net/http"
)

type TokenProvider func(r *http.Request) (string, error)
