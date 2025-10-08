package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	// KeelServiceTypeKey is the key for keel.service.type.
	KeelServiceTypeKey = attribute.Key("keel.service.type")
	// KeelServiceNameKey is the key for keel.service.name.
	KeelServiceNameKey = attribute.Key("keel.service.name")
	// KeelServiceInstKey is the key for keel.service.inst.
	KeelServiceInstKey = attribute.Key("keel.service.inst")
)

// KeelServiceType returns a new attribute.KeyValue for keel.service.type.
func KeelServiceType(v string) attribute.KeyValue {
	return KeelServiceTypeKey.String(v)
}

// KeelServiceName returns a new attribute.KeyValue for keel.service.name.
func KeelServiceName(v string) attribute.KeyValue {
	return KeelServiceNameKey.String(v)
}

// KeelServiceInst returns a new attribute.KeyValue for keel.service.inst.
func KeelServiceInst(v int) attribute.KeyValue {
	return KeelServiceInstKey.Int(v)
}
