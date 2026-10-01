// Package nats provides instrumented NATS and JetStream connection helpers
// for keel runtimes.
//
// [Connect] dials a NATS server with reconnect defaults, structured logging
// for connection lifecycle events and OpenTelemetry metrics for disconnects,
// reconnects and asynchronous errors. The connection is registered as a
// closer on the [github.com/foomo/keel.Runtime]. [NewJetStream] creates a
// JetStream context on top of such a connection with async publish defaults.
//
// # Usage
//
//	nc, err := nats.Connect(svr, "nats://localhost:4222")
//	if err != nil {
//		return err
//	}
//	js, err := nats.NewJetStream(svr, nc)
package nats
