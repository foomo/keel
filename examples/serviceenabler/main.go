package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/service"
)

func main() {
	svr := keel.NewServer()

	l := svr.Logger()
	c := svr.Config()

	enabled := config.GetBool(c, "service.enabled", true)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddServices(
		service.NewHTTP(l, "demo", "localhost:8080",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Set("service.enabled", !enabled())
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}),
		),
		keel.NewServiceEnabler(l, "service-enabler",
			func() keel.Service {
				return service.NewHTTP(l, "service", "localhost:8081", svs)
			},
			enabled,
		),
	)

	svr.Run()
}
