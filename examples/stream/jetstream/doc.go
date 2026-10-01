// Jetstream is an example service demonstrating publishing and subscribing to typed
// messages with the NATS JetStream stream.
//
// Start NATS with `docker run -it -p 4222:4222 --rm nats:2.7-alpine --jetstream`, run
// `go run ./examples/stream/jetstream`, then publish with `curl localhost:8080`.
package main
