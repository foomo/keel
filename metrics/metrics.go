package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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

func NewRequestDurationHistogram(namespace, subsystem string) prometheus.Histogram {
	return promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "The latency of the requests.",
		Buckets:   prometheus.ExponentialBuckets(.0001, 2, 50),
	})
}
