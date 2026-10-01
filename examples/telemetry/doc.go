// Telemetry is an example service demonstrating OpenTelemetry counters, up down
// counters and histograms alongside Prometheus exemplars, configured via OTEL_* env vars.
//
// Run it with e.g. `OTEL_METRICS_EXPORTER=console go run ./examples/telemetry`, then try
// `curl localhost:8080/count`, /up, /down or /histogram.
package main
