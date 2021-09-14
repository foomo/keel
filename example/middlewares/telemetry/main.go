package main

import (
	"net/http"

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

	svr.AddServices(keel.NewDefaultServiceHTTPPrometheus())

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Telemetry(),
		),
	)

	svr.Run()
}
