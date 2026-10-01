// Healthz is an example service demonstrating startup, readiness, liveness and
// always health probes.
//
// Run it with `go run ./examples/healthz`, then try `curl localhost:9400/healthz/readiness`
// (or /healthz, /healthz/liveness, /healthz/startup).
package main
