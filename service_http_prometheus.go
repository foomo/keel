package keel

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

const (
	DefaultServiceHTTPPrometheusAddr = ":9200"
	DefaultServiceHTTPPrometheusPath = "/metrics"
)

func NewServiceHTTPPrometheus(l *zap.Logger, addr, path string) *ServiceHTTP {
	handler := http.NewServeMux()
	handler.Handle(path, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	return NewServiceHTTP(l, addr, handler).SetName("prometheus")
}

func NewDefaultServiceHTTPPrometheus() *ServiceHTTP {
	return NewServiceHTTPPrometheus(
		log.Logger(),
		DefaultServiceHTTPPrometheusAddr,
		DefaultServiceHTTPPrometheusPath,
	)
}
