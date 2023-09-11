package service_test

import (
	"io"
	"net"
	"net/http"
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

func waitFor(addr string) {
	if _, err := net.DialTimeout("tcp", addr, 10*time.Second); err != nil {
		panic(err.Error())
	}
}

func httpGet(url string) string {
	resp, err := http.Get(url) //nolint:all
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}
