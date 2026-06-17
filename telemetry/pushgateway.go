package telemetry

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PushToGateway pushes the metrics gathered by prometheus.DefaultGatherer to a
// Prometheus Pushgateway at the given URL, grouped under the given job name.
//
// This is intended for short-lived workloads (e.g. Kubernetes Jobs) that exit
// before a scrape can occur. The OTEL Prometheus exporter registers into
// prometheus.DefaultRegisterer, so OTEL metrics are included in the push.
func PushToGateway(ctx context.Context, url, jobName string) error {
	return push.New(url, jobName).
		Gatherer(prometheus.DefaultGatherer).
		PushContext(ctx)
}
