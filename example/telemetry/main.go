package main

import (
	"math/rand"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
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
	//
	// OTEL_TRACE_RATIO="0.5"

	svr := keel.NewServer(
		keel.WithStdOutMeter(true),
		keel.WithStdOutTracer(true),
	)

	l := svr.Logger()

	meter := svr.Meter()

	// create demo service
	svs := http.NewServeMux()

	{ // counter
		counter, err := meter.SyncInt64().Counter(
			"a.counter",
			instrument.WithDescription("Count things"),
		)
		log.Must(l, err, "failed to create counter meter")

		svs.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
			counter.Add(r.Context(), 1, attribute.String("key", "value"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK!"))
		})
	}

	{ // up down
		upDown, err := meter.SyncInt64().UpDownCounter(
			"a.updown",
			instrument.WithDescription("Up down values"),
		)
		log.Must(l, err, "failed to create up down meter")

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
	}

	{ // histogram
		histogram, err := meter.SyncInt64().Histogram(
			"a.histogram",
			instrument.WithDescription("Up down values"),
		)
		log.Must(l, err, "failed to create up down meter")

		svs.HandleFunc("/histogram", func(w http.ResponseWriter, r *http.Request) {
			histogram.Record(r.Context(), int64(rand.Int()), attribute.String("key", "value"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK!"))
		})
	}

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.Telemetry(),
			middleware.Recover(),
		),
	)

	svr.Run()
}
