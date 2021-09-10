package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type (
	TelemetryOptions struct {
		Name     string
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
func Telemetry(opts ...TelemetryOption) Middleware {
	options := GetDefaultTelemetryOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return TelemetryWithOptions(options)
}

func TelemetryWithName(v string) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Name = v
	}
}

// TelemetryWithOtelOpts middleware options
func TelemetryWithOtelOpts(v ...otelhttp.Option) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.OtelOpts = v
	}
}

// TelemetryWithOptions middleware
func TelemetryWithOptions(opts TelemetryOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		if opts.Name != "" {
			name = opts.Name
		}
		return otelhttp.NewHandler(next, name, opts.OtelOpts...)
	}
}
