package main

import (
	"net/http"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/net/http/middleware"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	svs.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.ResponseTime(
				// automatically set cookie if not exists
				middleware.ResponseTimeWithMaxDuration(time.Millisecond*500),
			),
		),
	)

	svr.Run()
}
