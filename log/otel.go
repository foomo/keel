package log

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func Attributes(attrs ...attribute.KeyValue) []zap.Field {
	ret := make([]zap.Field, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Valid() {
			ret = append(ret, Attribute(attr))
		}
	}

	return ret
}

func Attribute(attr attribute.KeyValue) zap.Field {
	if !attr.Valid() {
		return zap.Skip()
	}
	switch attr.Value.Type() {
	case attribute.BOOL:
		return zap.Bool(AttributeKey(attr.Key), attr.Value.AsBool())
	case attribute.BOOLSLICE:
		return zap.Bools(AttributeKey(attr.Key), attr.Value.AsBoolSlice())
	case attribute.INT64:
		return zap.Int64(AttributeKey(attr.Key), attr.Value.AsInt64())
	case attribute.INT64SLICE:
		return zap.Int64s(AttributeKey(attr.Key), attr.Value.AsInt64Slice())
	case attribute.FLOAT64:
		return zap.Float64(AttributeKey(attr.Key), attr.Value.AsFloat64())
	case attribute.FLOAT64SLICE:
		return zap.Float64s(AttributeKey(attr.Key), attr.Value.AsFloat64Slice())
	case attribute.STRING:
		return zap.String(AttributeKey(attr.Key), attr.Value.AsString())
	case attribute.STRINGSLICE:
		if value := attr.Value.AsStringSlice(); len(value) == 1 {
			return zap.String(AttributeKey(attr.Key), value[0])
		} else {
			return zap.Strings(AttributeKey(attr.Key), value)
		}
	default:
		return zap.Any(AttributeKey(attr.Key), attr.Value.AsInterface())
	}
}

func AttributeKey(key attribute.Key) string {
	return strings.ReplaceAll(string(key), ".", "_")
}
