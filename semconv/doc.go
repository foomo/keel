// Package semconv provides OpenTelemetry attribute keys and constructors for
// keel specific semantic conventions.
//
// Each attribute is exposed as a
// [go.opentelemetry.io/otel/attribute.Key] constant (for example
// [KeelServiceNameKey]) and a constructor returning a
// [go.opentelemetry.io/otel/attribute.KeyValue] (for example
// [KeelServiceName]). Attributes cover keel services, debug state and gotsrpc
// calls.
package semconv
