package semconv

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

const (
	// NameKey is the key for name.
	NameKey = attribute.Key("name")
	// ValueKey is the key for value.
	ValueKey = attribute.Key("value")
	// NumKey is the key for num.
	NumKey = attribute.Key("num")
	// DurationKey is the key for duration, in seconds.
	DurationKey = attribute.Key("duration")
	// MaxDurationKey is the key for max_duration, in seconds.
	MaxDurationKey = attribute.Key("max_duration")
	// TraceIDKey is the key for trace.id.
	TraceIDKey = attribute.Key("trace.id")
	// SpanIDKey is the key for span.id.
	SpanIDKey = attribute.Key("span.id")
	// ContextErrorKey is the key for context.error.
	ContextErrorKey = attribute.Key("context.error")
)

// Name returns a new attribute.KeyValue for name.
func Name(v string) attribute.KeyValue {
	return NameKey.String(v)
}

// Value returns a new attribute.KeyValue for value with v formatted
// by %v.
func Value(v any) attribute.KeyValue {
	return ValueKey.String(fmt.Sprintf("%v", v))
}

// Num returns a new attribute.KeyValue for num.
func Num(v int) attribute.KeyValue {
	return NumKey.Int(v)
}

// Duration returns a new attribute.KeyValue for duration in seconds.
func Duration(v time.Duration) attribute.KeyValue {
	return DurationKey.Float64(v.Seconds())
}

// MaxDuration returns a new attribute.KeyValue for max_duration in seconds.
func MaxDuration(v time.Duration) attribute.KeyValue {
	return MaxDurationKey.Float64(v.Seconds())
}

// TraceID returns a new attribute.KeyValue for trace.id.
func TraceID(v string) attribute.KeyValue {
	return TraceIDKey.String(v)
}

// SpanID returns a new attribute.KeyValue for span.id.
func SpanID(v string) attribute.KeyValue {
	return SpanIDKey.String(v)
}

// ContextError returns a new attribute.KeyValue for context.error
// with the message of the given context error.
func ContextError(err error) attribute.KeyValue {
	return ContextErrorKey.String(err.Error())
}
