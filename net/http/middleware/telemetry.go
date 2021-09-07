package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	TelemetryOptions struct{}
	TelemetryOption  func(*TelemetryOptions)
)

// GetDefaultTelemetryOptions returns the default options
func GetDefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{}
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
		return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			start := time.Now()
			next.ServeHTTP(wr, r)
			duration := time.Since(start)

			log.WithHTTPRequest(l, r).Info(
				"handled http request",
				log.FDuration(duration),
				log.FHTTPStatusCode(wr.StatusCode()),
				log.FHTTPWroteBytes(int64(wr.Size())),
			)
		}), name)
	}
}
