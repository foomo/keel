package keel

import (
	"errors"
)

var (
	ErrServerNotRunning = errors.New("server not running")
	ErrServerShutdown   = errors.New("server is shutting down")
)
