package middleware

import (
	"net/http"

	"github.com/foomo/keel/net/http/context"
	"go.uber.org/zap"
)

type (
	RefererOptions struct {
		RequestHeader []string
		SetContext    bool
	}
	RefererOption func(*RefererOptions)
)

// GetDefaultRefererOptions returns the default options
func GetDefaultRefererOptions() RefererOptions {
	return RefererOptions{
		RequestHeader: []string{"X-Referer", "X-Referrer", "Referer", "Referrer"},
		SetContext:    true,
	}
}

// RefererWithRequestHeader middleware option
func RefererWithRequestHeader(v ...string) RefererOption {
	return func(o *RefererOptions) {
		o.RequestHeader = append(o.RequestHeader, v...)
	}
}

// RefererWithSetContext middleware option
func RefererWithSetContext(v bool) RefererOption {
	return func(o *RefererOptions) {
		o.SetContext = v
	}
}

// Referer middleware
func Referer(opts ...RefererOption) Middleware {
	options := GetDefaultRefererOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return RefererWithOptions(options)
}

// RefererWithOptions middleware
func RefererWithOptions(opts RefererOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var referer string
			for _, value := range opts.RequestHeader {
				if referer = r.Header.Get(value); referer != "" {
					break
				}
			}
			if referer != "" && opts.SetContext {
				r = r.WithContext(context.SetReferrer(r.Context(), referer))
			}
			next.ServeHTTP(w, r)
		})
	}
}
