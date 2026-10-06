package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	// DebugEnabledKey is the key for debug.enabled.
	//
	// Deprecated: use foomosemconv.SpanSamplingPriorityKey.
	DebugEnabledKey = attribute.Key("debug.enabled")
)

// DebugEnabled returns a new attribute.KeyValue for debug.enabled.
//
// Deprecated: use foomosemconv.SetSpanSamplingPriority.
func DebugEnabled(v bool) attribute.KeyValue {
	return DebugEnabledKey.Bool(v)
}
