package service_test

import (
	"syscall"
	"time"
)

// shutdown example after the given time
func shutdown(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
}
