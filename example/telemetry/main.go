package main

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/foomo/keel"
)

func main() {
	// Run this example with the following env vars:
	//
	// when otel is disabled (default: false) the stdout exporter is used
	// OTEL_ENABLED="false"
	//
	// name your service
	// OTEL_SERVICE_NAME="your-service-name"
	//
	// pretty print output (default: true)
	// OTEL_EXPORTER_STDOUT_PRETTY_PRINT="true"
	//
	// disable host metrics (default: true)
	// OTEL_METRICS_HOST_ENABLED="false"
	//
	// disable runtime metrics (default: true)
	// OTEL_METRICS_RUNTIME_ENABLED="false"

	svr := keel.NewServer(
		keel.WithStdOutMeter(false),
		keel.WithStdOutTracer(false),
	)

	l := svr.Logger()

	meter := svr.Meter()

	// create demo service
	svs := http.NewServeMux()

	counter := meter.NewInt64Counter(
		"a.counter",
		metric.WithDescription("Count things"),
	)

	svs.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		counter.Add(r.Context(), 1, attribute.String("key", "value"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK!"))
	})

	upDown := meter.NewInt64UpDownCounter(
		"a.updown",
		metric.WithDescription("Updown values"),
	)
	svs.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		upDown.Add(r.Context(), 1, attribute.String("key", "value"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK!"))
	})
	svs.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		upDown.Add(r.Context(), -1, attribute.String("key", "value"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK!"))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}
