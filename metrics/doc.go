// Package metrics provides constructors for Prometheus request metrics
// registered with the default registry via promauto.
//
// All constructors are deprecated in favor of OpenTelemetry instruments
// created from [github.com/foomo/keel/telemetry.Meter].
package metrics
