package service

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	DefaultHTTPPrometheusName = "prometheus"
	DefaultHTTPPrometheusAddr = ":9200"
	DefaultHTTPPrometheusPath = "/metrics"
)

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

func NewDefaultHTTPPrometheus(l *zap.Logger) *HTTP {
	return NewHTTPPrometheus(
		l,
		DefaultHTTPPrometheusName,
		DefaultHTTPPrometheusAddr,
		DefaultHTTPPrometheusPath,
	)
}
