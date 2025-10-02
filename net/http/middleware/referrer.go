package middleware

import (
	"net/http"
	"strings"

	"github.com/foomo/keel/net/http/context"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
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
		RequestHeader: []string{"X-Referer", "Referer"},
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
			span := trace.SpanFromContext(r.Context())
			span.AddEvent("Referer")

			var (
				key     string
				referer string
			)
			for _, value := range opts.RequestHeader {
				if referer = r.Header.Get(value); referer != "" {
					key = value
					break
				}
			}

			if referer != "" && opts.SetContext {
				span.SetAttributes(semconv.HTTPRequestHeader(strings.ToLower(key), referer))
				r = r.WithContext(context.SetReferer(r.Context(), referer))
			}

			next.ServeHTTP(w, r)
		})
	}
}
