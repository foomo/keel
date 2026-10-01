package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NewRequestSizeSummaryVec registers and returns a summary vector named
// <namespace>_<subsystem>_request_size_bytes that tracks request sizes,
// partitioned by labelNames. It panics if the metric is already registered.
//
// Deprecated: Use [github.com/foomo/keel/telemetry.Meter] instead.
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

// NewResponseSizeSummaryVec registers and returns a summary vector named
// <namespace>_<subsystem>_response_size_bytes that tracks response sizes,
// partitioned by labelNames. It panics if the metric is already registered.
//
// Deprecated: Use [github.com/foomo/keel/telemetry.Meter] instead.
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

// NewRequestsCounterVec registers and returns a counter vector named
// <namespace>_<subsystem>_requests_total that counts requests, partitioned
// by labelNames. It panics if the metric is already registered.
//
// Deprecated: Use [github.com/foomo/keel/telemetry.Meter] instead.
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

// NewRequestDurationHistogram registers and returns a histogram named
// <namespace>_<subsystem>_request_duration_seconds with 50 exponential
// buckets starting at 0.1ms. It panics if the metric is already registered.
//
// Deprecated: Use [github.com/foomo/keel/telemetry.Meter] instead.
func NewRequestDurationHistogram(namespace, subsystem string) prometheus.Histogram {
	return promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "The latency of the requests.",
		Buckets:   prometheus.ExponentialBuckets(.0001, 2, 50),
	})
}
