package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/net/http/middleware"
	"github.com/foomo/keel/service"
)

func main() {
	// Select the trace exporter via env, e.g. OTEL_TRACES_EXPORTER="console".
	svr := keel.NewServer(
		keel.WithTelemetry(),
	)

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs,
			middleware.Telemetry(
				middleware.TelemetryWithInjectPropagationHeader(true),
			),
		),
	)

	svr.Run()
}
