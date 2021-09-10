package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Deprecated: NewRequestSizeSummaryVec use telemetry.Meter(...)
func NewRequestSizeSummaryVec(namespace, subsystem string, labelNames []string) *prometheus.SummaryVec {
	return promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "Tracks the size of HTTP requests.",
		},
		labelNames,
	)
}

// Deprecated: NewResponseSizeSummaryVec use telemetry.Meter(...)
func NewResponseSizeSummaryVec(namespace, subsystem string, labelNames []string) *prometheus.SummaryVec {
	return promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "Tracks the size of responses.",
		},
		labelNames,
	)
}

// Deprecated: NewRequestsCounterVec use telemetry.Meter(...)
func NewRequestsCounterVec(namespace, subsystem string, labelNames []string) *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Tracks the number of requests.",
		},
		labelNames,
	)
}

// Deprecated: NewRequestDurationHistogram use telemetry.Meter(...)
func NewRequestDurationHistogram(namespace, subsystem string) prometheus.Histogram {
	return promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "The latency of the requests.",
		Buckets:   prometheus.ExponentialBuckets(.0001, 2, 50),
	})
}
