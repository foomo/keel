package service

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Default name, address and path of the service returned by
// [NewDefaultHTTPPrometheus].
var (
	DefaultHTTPPrometheusName = "prometheus"
	DefaultHTTPPrometheusAddr = ":9200"
	DefaultHTTPPrometheusPath = "/metrics"
)

// NewHTTPPrometheus returns an [HTTP] service exposing the metrics of
// [prometheus.DefaultGatherer] on path, with OpenMetrics enabled.
func NewHTTPPrometheus(l *zap.Logger, name, addr, path string) *HTTP {
	handler := http.NewServeMux()
	handler.Handle(path, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	return NewHTTP(l, name, addr, handler)
}

// NewDefaultHTTPPrometheus returns [NewHTTPPrometheus] using
// [DefaultHTTPPrometheusName], [DefaultHTTPPrometheusAddr] and
// [DefaultHTTPPrometheusPath].
func NewDefaultHTTPPrometheus(l *zap.Logger) *HTTP {
	return NewHTTPPrometheus(
		l,
		DefaultHTTPPrometheusName,
		DefaultHTTPPrometheusAddr,
		DefaultHTTPPrometheusPath,
	)
}
