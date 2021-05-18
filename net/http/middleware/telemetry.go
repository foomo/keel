package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/metrics"
)

type TelemetryConfig struct {
	ServiceName string
}

var DefaultTelemetryConfig = TelemetryConfig{
	ServiceName: "service",
}

func Telemetry() Middleware {
	return TelemetryWithConfig(DefaultTelemetryConfig)
}

func TelemetryWithConfig(config TelemetryConfig) Middleware {
	if config.ServiceName == "" {
		config.ServiceName = DefaultTelemetryConfig.ServiceName
	}

	requestSize := metrics.NewHTTPRequestSizeSummaryVec(config.ServiceName)
	responseSize := metrics.NewHTTPResponseSizeSummaryVec(config.ServiceName)
	requestDuration := metrics.NewHTTPRequestDurationHistogram(config.ServiceName)
	requestsCounter := metrics.NewHTTPRequestsCounterVec(config.ServiceName)

	return func(l *zap.Logger, next http.Handler) http.Handler {
		return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// wrap response write to get access to status & size
			wr := wrapResponseWriter(w)

			next.ServeHTTP(wr, r)

			if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
				requestDuration.(prometheus.ExemplarObserver).ObserveWithExemplar(
					time.Since(start).Seconds(),
					prometheus.Labels{"traceID": spanCtx.TraceID().String()},
				)
			} else {
				requestDuration.Observe(time.Since(start).Seconds())
			}
			requestSize.WithLabelValues(strings.ToLower(r.Method), wr.Status()).Observe(float64(r.ContentLength))
			responseSize.WithLabelValues(strings.ToLower(r.Method), wr.Status()).Observe(float64(wr.Size()))
			requestsCounter.WithLabelValues(strings.ToLower(r.Method), wr.Status()).Inc()
		}), config.ServiceName)
	}
}
