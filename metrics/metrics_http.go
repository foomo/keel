package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// NewHTTPRequestSizeSummaryVec returns [NewRequestSizeSummaryVec] with
// subsystem "http" and labels "method" and "code".
//
// Deprecated: Use otelhttp instrumentation instead.
func NewHTTPRequestSizeSummaryVec(namespace string) *prometheus.SummaryVec {
	return NewRequestSizeSummaryVec(namespace, "http", []string{"method", "code"})
}

// NewHTTPResponseSizeSummaryVec returns [NewResponseSizeSummaryVec] with
// subsystem "http" and labels "method" and "code".
//
// Deprecated: Use otelhttp instrumentation instead.
func NewHTTPResponseSizeSummaryVec(namespace string) *prometheus.SummaryVec {
	return NewResponseSizeSummaryVec(namespace, "http", []string{"method", "code"})
}

// NewHTTPRequestsCounterVec returns [NewRequestsCounterVec] with
// subsystem "http" and labels "method" and "code".
//
// Deprecated: Use otelhttp instrumentation instead.
func NewHTTPRequestsCounterVec(namespace string) *prometheus.CounterVec {
	return NewRequestsCounterVec(namespace, "http", []string{"method", "code"})
}

// NewHTTPRequestDurationHistogram returns [NewRequestDurationHistogram]
// with subsystem "http".
//
// Deprecated: Use otelhttp instrumentation instead.
func NewHTTPRequestDurationHistogram(namespace string) prometheus.Histogram {
	return NewRequestDurationHistogram(namespace, "http")
}
