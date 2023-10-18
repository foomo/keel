package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/net/http/middleware"
	"github.com/foomo/keel/service"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic("handled")
	})
	svs.HandleFunc("/abort", func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs,
			middleware.Recover(
				middleware.RecoverWithDisablePrintStack(true),
			),
		),
	)

	svr.Run()
}
