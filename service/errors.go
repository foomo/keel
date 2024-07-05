package service

import (
	"errors"
)

var (
	ErrServiceNotRunning = errors.New("service not running")
	ErrServiceShutdown   = errors.New("service shutdown")
)
