package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewHTTPRequestSizeSummaryVec(namespace string) *prometheus.SummaryVec {
	return NewRequestSizeSummaryVec(namespace, "http", []string{"method", "code"})
}

func NewHTTPResponseSizeSummaryVec(namespace string) *prometheus.SummaryVec {
	return NewResponseSizeSummaryVec(namespace, "http", []string{"method", "code"})
}

func NewHTTPRequestsCounterVec(namespace string) *prometheus.CounterVec {
	return NewRequestsCounterVec(namespace, "http", []string{"method", "code"})
}

func NewHTTPRequestDurationHistogram(namespace string) prometheus.Histogram {
	return NewRequestDurationHistogram(namespace, "http")
}
