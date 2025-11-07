package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	ProfileNameKey = attribute.Key("profile.name")
)

// ProfileName returns a new attribute.KeyValue for tracking.id.
func ProfileName(v string) attribute.KeyValue {
	return ProfileNameKey.String(v)
}
