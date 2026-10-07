// Package semconv provides OpenTelemetry attribute keys and constructors for
// keel specific semantic conventions.
//
// Each attribute is exposed as a
// [go.opentelemetry.io/otel/attribute.Key] constant (for example
// [KeelServiceNameKey]) and a constructor returning a
// [go.opentelemetry.io/otel/attribute.KeyValue] (for example
// [KeelServiceName]). Attributes cover keel services, jobs, closers, circuit
// breakers, RBAC, streams, debug state and generic values (name,
// value, duration, trace and span ids) that replace the legacy log fields.
// Use them with log.Attribute and log.Attributes to share keys between logs,
// metrics and traces. Prefer upstream go.opentelemetry.io/otel/semconv keys
// where they exist.
package semconv
