package keel

import (
	"errors"
)

var (
	ErrServerNotRunning  = errors.New("server not running")
	ErrServiceNotRunning = errors.New("service not running")
)
