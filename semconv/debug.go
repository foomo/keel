package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	// DebugEnabledKey is the key for debug.enabled.
	DebugEnabledKey = attribute.Key("debug.enabled")
)

// DebugEnabled returns a new attribute.KeyValue for debug.enabled.
func DebugEnabled(v bool) attribute.KeyValue {
	return DebugEnabledKey.Bool(v)
}
