package main

import (
	"net/http"
	"os"

	"github.com/foomo/keel"
	"github.com/foomo/keel/example/probes/handler"
)

func main() {
	// you can override the below config by settings env vars
	_ = os.Setenv("SERVICE_HEALTHZ_ENABLED", "true")

	svr := keel.NewServer(
		// add probes service listening on 0.0.0.0:9400
		// allows you to use probes for health checks in cluster: GET 0.0.0.0:9400/healthz
		keel.WithHTTPProbesService(false),
	)

	l := svr.Logger()

	// alternatively you can add them manually
	svr.AddServices(keel.NewDefaultServiceHTTPZap())

	h := handler.New(l)
	// Add probe handlers
	svr.AddLivelinessProbes(h)
	// svr.AddReadinessProbes(h)
	// svr.AddStartupProbes(h)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs),
	)

	svr.Run()
}
