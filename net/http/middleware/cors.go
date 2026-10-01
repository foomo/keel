package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
)

type (
	// CORSOptions configures the [CORS] middleware.
	CORSOptions struct {
		// AllowOrigins lists the allowed origins. "*" allows any origin. Entries
		// may contain the wildcards "*" (any sequence) and "?" (any single
		// character), e.g. "https://*.example.com".
		AllowOrigins []string
		// AllowMethods lists the methods sent in Access-Control-Allow-Methods
		// on preflight responses.
		AllowMethods []string
		// AllowHeaders lists the headers sent in Access-Control-Allow-Headers on
		// preflight responses. When empty, the request's
		// Access-Control-Request-Headers value is echoed.
		AllowHeaders []string
		// AllowCredentials reports whether Access-Control-Allow-Credentials is
		// sent. When set together with the "*" origin, the request origin is
		// echoed instead of "*".
		AllowCredentials bool
		// ExposeHeaders lists the headers sent in Access-Control-Expose-Headers
		// on non-preflight responses.
		ExposeHeaders []string
		// MaxAge is the Access-Control-Max-Age in seconds sent on preflight
		// responses. Zero omits the header.
		MaxAge int
	}
	// CORSOption configures [CORSOptions].
	CORSOption func(*CORSOptions)
)

// GetDefaultCORSOptions returns the default options, which allow any origin
// and the methods GET, HEAD, PUT, PATCH, POST and DELETE.
func GetDefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}
}

// CORSWithAllowOrigins sets [CORSOptions.AllowOrigins].
func CORSWithAllowOrigins(v ...string) CORSOption {
	return func(o *CORSOptions) {
		o.AllowOrigins = v
	}
}

// CORSWithAllowMethods sets [CORSOptions.AllowMethods].
func CORSWithAllowMethods(v ...string) CORSOption {
	return func(o *CORSOptions) {
		o.AllowMethods = v
	}
}

// CORSWithAllowHeaders sets [CORSOptions.AllowHeaders].
func CORSWithAllowHeaders(v ...string) CORSOption {
	return func(o *CORSOptions) {
		o.AllowHeaders = v
	}
}

// CORSWithAllowCredentials sets [CORSOptions.AllowCredentials].
func CORSWithAllowCredentials(v bool) CORSOption {
	return func(o *CORSOptions) {
		o.AllowCredentials = v
	}
}

// CORSWithExposeHeaders sets [CORSOptions.ExposeHeaders].
func CORSWithExposeHeaders(v ...string) CORSOption {
	return func(o *CORSOptions) {
		o.ExposeHeaders = v
	}
}

// CORSWithMaxAge sets [CORSOptions.MaxAge].
func CORSWithMaxAge(v int) CORSOption {
	return func(o *CORSOptions) {
		o.MaxAge = v
	}
}

// CORS returns a middleware that handles cross-origin resource sharing.
// Preflight (OPTIONS) requests are answered with 204 No Content and never
// reach the next handler, regardless of whether the origin is allowed. Other
// requests are always passed on and receive the CORS response headers only
// when their origin is allowed.
func CORS(opts ...CORSOption) keelhttp.Middleware {
	options := GetDefaultCORSOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return CORSWithOptions(options)
}

// CORSWithOptions is like [CORS] but takes fully populated options.
func CORSWithOptions(opts CORSOptions) keelhttp.Middleware {
	allowOriginPatterns := make([]*regexp.Regexp, len(opts.AllowOrigins))
	for i, origin := range opts.AllowOrigins {
		pattern := regexp.QuoteMeta(origin)
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
		pattern = "^" + pattern + "$"
		allowOriginPatterns[i] = regexp.MustCompile(pattern)
	}

	allowMethods := strings.Join(opts.AllowMethods, ",")
	allowHeaders := strings.Join(opts.AllowHeaders, ",")
	exposeHeaders := strings.Join(opts.ExposeHeaders, ",")
	maxAge := strconv.Itoa(opts.MaxAge)

	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("CORS")
			}

			origin := r.Header.Get(keelhttp.HeaderOrigin)
			allowOrigin := ""

			w.Header().Add(keelhttp.HeaderVary, keelhttp.HeaderOrigin)

			preflight := r.Method == http.MethodOptions

			// No Origin provided
			if origin == "" {
				if !preflight {
					next.ServeHTTP(w, r)
					return
				}

				w.WriteHeader(http.StatusNoContent)

				return
			}

			// Check allowed origins
			for _, value := range opts.AllowOrigins {
				if value == "*" && opts.AllowCredentials {
					allowOrigin = origin
					break
				}

				if value == "*" || value == origin {
					allowOrigin = value
					break
				}

				if matchSubdomain(origin, value) {
					allowOrigin = origin
					break
				}
			}

			// Check allowed origin patterns
			for _, re := range allowOriginPatterns {
				if allowOrigin == "" {
					_, after, found := strings.Cut(origin, "://")
					if !found {
						continue
					}

					if len(after) > 253 {
						break
					}

					if re.MatchString(origin) {
						allowOrigin = origin
						break
					}
				}
			}

			// Origin not allowed
			if allowOrigin == "" && !preflight {
				next.ServeHTTP(w, r)
				return
			} else if allowOrigin == "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Simple request
			if !preflight {
				w.Header().Set(keelhttp.HeaderAccessControlAllowOrigin, allowOrigin)

				if opts.AllowCredentials {
					w.Header().Set(keelhttp.HeaderAccessControlAllowCredentials, "true")
				}

				if exposeHeaders != "" {
					w.Header().Set(keelhttp.HeaderAccessControlExposeHeaders, exposeHeaders)
				}

				next.ServeHTTP(w, r)

				return
			}

			// Preflight request
			w.Header().Add(keelhttp.HeaderVary, keelhttp.HeaderAccessControlRequestMethod)
			w.Header().Add(keelhttp.HeaderVary, keelhttp.HeaderAccessControlRequestHeaders)
			w.Header().Set(keelhttp.HeaderAccessControlAllowOrigin, allowOrigin)
			w.Header().Set(keelhttp.HeaderAccessControlAllowMethods, allowMethods)

			if opts.AllowCredentials {
				w.Header().Set(keelhttp.HeaderAccessControlAllowCredentials, "true")
			}

			if allowHeaders != "" {
				w.Header().Set(keelhttp.HeaderAccessControlAllowHeaders, allowHeaders)
			} else if h := r.Header.Get(keelhttp.HeaderAccessControlRequestHeaders); h != "" {
				w.Header().Set(keelhttp.HeaderAccessControlAllowHeaders, h)
			}

			if opts.MaxAge > 0 {
				w.Header().Set(keelhttp.HeaderAccessControlMaxAge, maxAge)
			}

			w.WriteHeader(http.StatusNoContent)
		})
	}
}

// matchScheme reports whether domain and pattern have the same scheme.
func matchScheme(domain, pattern string) bool {
	domScheme, _, domFound := strings.Cut(domain, ":")
	patScheme, _, patFound := strings.Cut(pattern, ":")

	return domFound && patFound && domScheme == patScheme
}

// matchSubdomain reports whether domain matches pattern, comparing the
// dot-separated authority components right to left. A "*" component in
// pattern matches the remaining components, e.g. "https://*.example.com"
// matches "https://api.example.com".
func matchSubdomain(domain, pattern string) bool {
	if !matchScheme(domain, pattern) {
		return false
	}

	_, domAuth, domFound := strings.Cut(domain, "://")
	if !domFound {
		return false
	}

	_, patAuth, patFound := strings.Cut(pattern, "://")
	if !patFound {
		return false
	}

	// to avoid long loop by invalid long domain
	if len(domAuth) > 253 {
		return false
	}

	// Compare components right-to-left without allocating slices.
	// Walk backward through both strings, extracting one dot-delimited
	// component at a time and comparing them.
	domPos := len(domAuth)
	patPos := len(patAuth)

	for domPos > 0 && patPos > 0 {
		// Extract rightmost remaining component from domain.
		domDot := strings.LastIndexByte(domAuth[:domPos], '.')
		domSeg := domAuth[domDot+1 : domPos]
		domPos = domDot // -1 when no more dots, which terminates next iteration

		// Extract rightmost remaining component from pattern.
		patDot := strings.LastIndexByte(patAuth[:patPos], '.')
		patSeg := patAuth[patDot+1 : patPos]
		patPos = patDot

		if patSeg == "*" {
			return true
		}

		if patSeg != domSeg {
			return false
		}
	}

	return false
}
