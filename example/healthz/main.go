package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/example/healthz/handler"
)

func main() {
	// you can override the below config by settings env vars
	_ = os.Setenv("SERVICE_HEALTHZ_ENABLED", "true")

	svr := keel.NewServer(
		// add probes service listening on :9400
		// allows you to use probes for health checks in cluster: GET :9400/healthz
		keel.WithHTTPHealthzService(true),
	)

	l := svr.Logger()

	// Add probe handlers
	svr.AddAlwaysHealthzers(handler.New(l, "always"))
	svr.AddStartupHealthzers(handler.New(l, "startup"))
	svr.AddLivenessHealthzers(handler.New(l, "liveness"))
	svr.AddReadinessHealthzers(handler.New(l, "readiness"))

	svr.AddAlwaysHealthzers(keel.NewHealthzerFn(func(ctx context.Context) error {
		l.Info("healther fn")
		return nil
	}))

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	l.Info("doing some initialization")
	time.Sleep(10 * time.Second)
	l.Info("initialization done")

	// TODO wait for services to be started
	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}
