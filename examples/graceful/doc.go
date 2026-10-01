// Graceful is an example service demonstrating graceful shutdown with readiness
// probes and registered closers.
//
// Run it with `go run ./examples/graceful`. It polls localhost:8080 and the
// readiness probe on localhost:9400, then sends itself SIGTERM and logs the shutdown sequence.
package main
