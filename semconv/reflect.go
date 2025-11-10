package semconv

import (
	"reflect"

	"go.opentelemetry.io/otel/attribute"
)

const (
	// RefectTypeKey is the key for reflect.type.
	RefectTypeKey = attribute.Key("reflect.type")
)

func RefectType(v any) attribute.KeyValue {
	return RefectTypeKey.String(reflect.TypeOf(v).String())
}
