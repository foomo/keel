package service

import (
	"github.com/pkg/errors"
)

var (
	ErrServiceNotRunning = errors.New("service not running")
	ErrServiceShutdown   = errors.New("service shutdown")
)
