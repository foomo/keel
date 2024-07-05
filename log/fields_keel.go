package log

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	KeelServiceTypeKey = attribute.Key("keel.service.type")
	KeelServiceNameKey = attribute.Key("keel.service.name")
	KeelServiceInstKey = attribute.Key("keel.service.inst")
)
