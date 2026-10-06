// Package log provides the zap logger setup used by keel together with
// helpers for building structured log fields.
//
// On initialization the package installs a zap global logger configured from
// the LOG_LEVEL (default "info") and LOG_FORMAT ("json" or "console", default
// "json") environment variables; see [NewLogger]. [Logger] returns it and
// [AtomicLevel] allows changing its level at runtime.
//
// # Fields
//
// [Attribute] and [Attributes] convert OpenTelemetry attributes to fields, so
// the same semconv attributes can be shared between logs, metrics and traces.
// Use the upstream go.opentelemetry.io/otel/semconv package or
// github.com/foomo/keel/semconv for keel specific keys. With* functions return
// a child logger with fields added, e.g. from an [net/http.Request]:
//
//	l := log.WithHTTPRequest(nil, r)
//	l.Info("handled", log.Attribute(keelsemconv.Duration(time.Since(start))))
//
// F* functions create single [zap.Field] values with legacy keys and are kept
// for backward compatibility.
//
// TODO: deprecate the remaining F* functions and *Key constants in the next
// version in favor of semconv attributes with [Attribute] and [Attributes].
//
// A [Labeler] stored in a context collects fields that are added while a
// request is processed.
package log
