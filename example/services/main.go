package main

import (
	"net/http"
	"os"

	"github.com/foomo/keel"
)

// Probe handler
// type PingHandler struct {
// 	l *zap.Logger
// }

func main() {
	// you can override the below config by settings env vars
	_ = os.Setenv("SERVICE_ZAP_ENABLED", "true")
	_ = os.Setenv("SERVICE_VIPER_ENABLED", "true")
	_ = os.Setenv("SERVICE_PROMETHEUS_ENABLED", "true")

	svr := keel.NewServer(
		// add zap service listening on localhost:9100
		// allows you to view / change the log level: GET / PUT localhost:9100/log
		keel.WithHTTPZapService(false),
		// add viper service listening on localhost:9300
		// allows you to view / change the configuration: GET / PUT localhost:9300/config
		keel.WithHTTPViperService(false),
		// add prometheus service listening on 0.0.0.0:9200
		// allows you to collect prometheus metrics: GET 0.0.0.0:9200/metrics
		keel.WithHTTPPrometheusService(false),
		// add probes service listening on 0.0.0.0:9400
		// allows you to use probes for health checks in cluster: GET 0.0.0.0:9400/healthz
		keel.WithHTTPProbesService(false),
	)

	l := svr.Logger()

	// alternatively you can add them manually
	// svr.AddServices(keel.NewDefaultServiceHTTPZap())
	// svr.AddServices(keel.NewDefaultServiceHTTPViper())

	// Add probe handelers
	// svr.AddProbeHandlers(&PingHandler{l: l}, keel.Liveliness)
	// svr.AddProbeHandlers(&PingHandler{l: l}, keel.Readiness)
	// svr.AddProbeHandlers(&PingHandler{l: l}, keel.Startup)

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

// Probe handler ping function
// func (p *PingHandler) Ping() bool {
// 	log.WithServiceName(p.l, "SERVICE").Error("FAILED")
// 	return true
// }
