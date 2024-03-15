package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/examples/healthz/handler"
	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/service"
)

// See k8s for probe documentation
// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#types-of-probe
func main() {
	service.DefaultHTTPHealthzAddr = "localhost:9400"

	// you can override the below config by settings env vars
	_ = os.Setenv("SERVICE_HEALTHZ_ENABLED", "true")

	svr := keel.NewServer(
		// allows you to use probes for health checks in cluster:
		//	GET :9400/healthz
		//  GET :9400/healthz/readiness
		//  GET :9400/healthz/liveness
		//  GET :9400/healthz/startup
		keel.WithHTTPHealthzService(true),
	)

	l := svr.Logger()

	// Add a probe that always be called
	ah := handler.New(l, "always")
	svr.AddAlwaysHealthzers(ah)

	// Add a probe only on startup
	sh := handler.New(l, "startup")
	svr.AddStartupHealthzers(sh)

	// Add a probe only on liveness
	lh := handler.New(l, "liveness")
	svr.AddLivenessHealthzers(lh)

	// Add a probe only on readiness
	rh := handler.New(l, "readiness")
	svr.AddReadinessHealthzers(rh)

	// add inline probe e.g. in case you start go routines
	svr.AddAlwaysHealthzers(healthz.NewHealthzerFn(func(ctx context.Context) error {
		l.Info("healther fn")
		return nil
	}))

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// long taking initialization
	l.Info("doing some initialization")
	select {
	case <-time.After(10 * time.Second):
		l.Info("initialization done")
	case <-svr.CancelContext().Done():
		l.Info("initialization canceled")
	}

	// add services
	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs),
	)

	// start serer
	svr.Run()
}
