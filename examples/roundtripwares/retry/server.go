package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
)

func server() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()

	counter := 0
	svs.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		counter++
		if counter < 10 {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}
