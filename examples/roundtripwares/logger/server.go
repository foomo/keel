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

	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	svs.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}
