package middleware

import (
	"net/http"
)

type Skipper func(*http.Request) bool
