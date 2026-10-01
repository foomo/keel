// Package service provides an embedded NATS server that runs as a keel
// service.
//
// [EmbeddedServer] starts an in-process NATS server in its Start method and
// blocks until the server shuts down, which makes it suitable for local
// development and tests.
package service
