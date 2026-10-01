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
// F* functions create single [zap.Field] values with well-known keys, and
// [Attribute] and [Attributes] convert OpenTelemetry attributes to fields.
// With* functions return a child logger with fields added, e.g. from an
// [net/http.Request]:
//
//	l := log.WithHTTPRequest(nil, r)
//	l.Info("handled", log.FDuration(time.Since(start)))
//
// A [Labeler] stored in a context collects fields that are added while a
// request is processed.
package log
