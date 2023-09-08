package service_test

import (
	"syscall"
	"time"
)

// shutdown example after the given time
func shutdownAfter(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		shutdown()
	}()
}

func shutdown() {
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		panic(err)
	}
}
