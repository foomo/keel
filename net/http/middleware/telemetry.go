package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
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

			// wrap response write to get access to status & size
			wr := wrapResponseWriter(w)

			start := time.Now()
			next.ServeHTTP(wr, r)
			duration := time.Since(start)

			if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
				requestDuration.(prometheus.ExemplarObserver).ObserveWithExemplar(
					duration.Seconds(),
					prometheus.Labels{"traceID": spanCtx.TraceID().String()},
				)
			} else {
				requestDuration.Observe(duration.Seconds())
			}
			requestSize.WithLabelValues(strings.ToLower(r.Method), wr.Status()).Observe(float64(r.ContentLength))
			responseSize.WithLabelValues(strings.ToLower(r.Method), wr.Status()).Observe(float64(wr.Size()))
			requestsCounter.WithLabelValues(strings.ToLower(r.Method), wr.Status()).Inc()

			log.WithHTTPRequest(l, r).Info(
				"handled http request",
				log.FDuration(time.Since(start)),
				log.FHTTPStatusCode(wr.StatusCode()),
				log.FHTTPWroteBytes(int64(wr.Size())),
			)
		}), config.ServiceName)
	}
}
