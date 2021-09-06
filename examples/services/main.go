package main

import (
	"net/http"
	"os"

	"github.com/foomo/keel"
)

func ExampleServices() {
	// you can override the below config by settings env vars
	_ = os.Setenv("SERVICE_ZAP_ENABLED", "false")
	_ = os.Setenv("SERVICE_VIPER_ENABLED", "false")
	_ = os.Setenv("SERVICE_PROMETHEUS_ENABLED", "false")

	svr := keel.NewServer(
		// add zap service listening on localhost:9100
		// allows you to view / change the log level: GET / PUT localhost:9100/log
		keel.WithHTTPZapService(true),
		// add viper service listening on localhost:9300
		// allows you to view / change the configuration: GET / PUT localhost:9300/config
		keel.WithHTTPViperService(true),
		// add prometheus service listening on 0.0.0.0:9200
		// allows you to collect prometheus metrics: GET 0.0.0.0:9200/metrics
		keel.WithHTTPPrometheusService(true),
	)

	l := svr.Logger()

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

func ExampleCustomServices() {
	svr := keel.NewServer()

	l := svr.Logger()

	// add zap service listening on localhost:9100
	// allows you to view / change the log level: GET / PUT localhost:9100/log
	svr.AddServices(keel.NewDefaultServiceHTTPZap())

	// add viper service listening on localhost:9300
	// allows you to view / change the configuration: GET / PUT localhost:9300/config
	svr.AddServices(keel.NewDefaultServiceHTTPViper())

	// add prometheus service listening on 0.0.0.0:9200
	// allows you to collect prometheus metrics: GET 0.0.0.0:9200/metrics
	svr.AddServices(keel.NewDefaultServiceHTTPPrometheus())

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
