package main

import (
	"os"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
)

func main() {
	service.DefaultHTTPPrometheusAddr = "localhost:9200"

	// you can override the below config by settings env vars
	_ = os.Setenv("SERVICE_ZAP_ENABLED", "true")
	_ = os.Setenv("SERVICE_VIPER_ENABLED", "true")
	_ = os.Setenv("SERVICE_PPROF_ENABLED", "true")
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
		// add go pprof service listening on 0.0.0.0:6060
		// allows you to use go tool pprof: GET 0.0.0.0:6060/debug/pprof
		keel.WithHTTPPProfService(false),
	)

	// alternatively you can add them manually
	// svr.AddServices(keel.NewDefaultServiceHTTPZap())
	// svr.AddServices(keel.NewDefaultServiceHTTPViper())
	// svr.AddServices(keel.NewDefaultServiceHTTPPProf())
	// svr.AddServices(keel.NewDefaultServiceHTTPPrometheus())

	svr.Run()
}
