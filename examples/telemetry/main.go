package main

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/foomo/keel"
	"github.com/foomo/keel/telemetry"
)

func main() {
	// Run this example with the following env vars:
	//
	// name your serivce
	// OTEL_SERVICE_NAME="your-service-name"
	//
	// when otel is disabled (default: false) the stdout exporter is used
	// OTEL_ENABLED="false"
	//
	// enable metrics output (default: false)
	// OTEL_EXPORTER_STDOUT_METRICS_ENABLED="true"
	//
	// enable trace output (default: false)
	// OTEL_EXPORTER_STDOUT_TRACE_ENABLED="true"
	//
	// pretty print ouput (default: true)
	// OTEL_EXPORTER_STDOUT_PRETTY_PRINT="true"
	//
	// disable host metrics (default: true)
	// OTEL_METRICS_HOST_ENABLED="false"
	//
	// disable runtime metrics (default: true)
	// OTEL_METRICS_RUNTIME_ENABLED="false"

	svr := keel.NewServer()

	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		meter := telemetry.Meter("demo")

		counter := meter.NewInt64Counter(
			"a.counter",
			metric.WithDescription("Count things"),
		)

		recorder := meter.NewInt64ValueRecorder(
			"a.valuerecorder",
			metric.WithDescription("Records values"),
		)

		updown := meter.NewInt64UpDownCounter(
			"a.updown",
			metric.WithDescription("Updown values"),
		)

		counter.Add(r.Context(), 100, attribute.String("key", "value"))
		counter.Add(r.Context(), 100, attribute.String("key", "value"))

		recorder.Record(r.Context(), 100, attribute.String("key", "value"))

		updown.Add(r.Context(), 120, attribute.String("key", "value"))
		updown.Add(r.Context(), 10, attribute.String("key", "value"))
		updown.Add(r.Context(), -10, attribute.String("key", "value"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK!"))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs),
	)

	svr.Run()
}
