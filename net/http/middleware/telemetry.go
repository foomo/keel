package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type (
	TelemetryOptions struct {
		OtelOpts []otelhttp.Option
	}
	TelemetryOption func(*TelemetryOptions)
)

// GetDefaultTelemetryOptions returns the default options
func GetDefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		OtelOpts: []otelhttp.Option{},
	}
}

// Telemetry middleware
func Telemetry(name string, opts ...TelemetryOption) Middleware {
	options := GetDefaultTelemetryOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return TelemetryWithOptions(name, options)
}

// TelemetryWithOptions middleware
func TelemetryWithOptions(name string, opts TelemetryOptions) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, name, opts.OtelOpts...)
	}
}
