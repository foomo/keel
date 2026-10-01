// Jetstreamraw is an example service demonstrating raw publishing and subscribing
// through the underlying NATS JetStream context.
//
// Start NATS with `docker run -it -p 4222:4222 --rm nats:2.7-alpine --jetstream`, run
// `go run ./examples/stream/jetstreamraw`, then publish with `curl localhost:8080`.
package main
