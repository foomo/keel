package keel

import (
	"errors"
)

var (
	// ErrServerNotRunning is returned by [Server.Healthz] while the server is not running.
	ErrServerNotRunning = errors.New("server not running")
	// ErrServerShutdown is returned by the readiness probe and the server's error
	// group once graceful shutdown has started.
	ErrServerShutdown = errors.New("server is shutting down")
)
