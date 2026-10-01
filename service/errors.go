package service

import (
	"errors"
)

var (
	// ErrServiceNotRunning is returned by the Healthz methods while a service is
	// not running.
	ErrServiceNotRunning = errors.New("service not running")
	// ErrServiceShutdown is the cancellation cause of a service context canceled
	// by Close.
	ErrServiceShutdown = errors.New("service shutdown")
)
