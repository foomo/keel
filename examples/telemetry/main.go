package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/foomo/keel/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var metricRequestLatency = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: "demo",
	Name:      "request_latency_seconds",
	Help:      "Request Latency",
	Buckets:   prometheus.ExponentialBuckets(.0001, 2, 50),
})

func main() {
	// Run this example with the following env vars:
	//
	// select the exporter per signal (none(default) | console | otlp | prometheus):
	// OTEL_TRACES_EXPORTER="console"
	// OTEL_METRICS_EXPORTER="console"
	// OTEL_LOGS_EXPORTER="console"
	//
	// for otlp, pick the protocol (grpc | http/protobuf(default)):
	// OTEL_EXPORTER_OTLP_PROTOCOL="grpc"
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
		keel.WithTelemetry(),
	)

	l := svr.Logger()

	meter := svr.Meter()

	// create demo service
	svs := http.NewServeMux()

	{ // counter
		counter, err := meter.Int64Counter(
			"demo.count",
			metric.WithDescription("Number of calls to the count endpoint."),
			metric.WithUnit("{call}"),
		)
		log.Must(l, err, "failed to create counter meter")

		svs.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
			counter.Add(r.Context(), 1, metric.WithAttributes(attribute.String("key", "value")))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK!"))
		})
	}

	promauto.NewCounter(prometheus.CounterOpts{
		Namespace:   "foo",
		Subsystem:   "",
		Name:        "bar",
		Help:        "blubb",
		ConstLabels: nil,
	})

	{ // up down
		upDown, err := meter.Int64UpDownCounter(
			"demo.level",
			metric.WithDescription("Current level raised by the up and lowered by the down endpoint."),
			metric.WithUnit("{level}"),
		)
		log.Must(l, err, "failed to create up down meter")

		svs.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
			upDown.Add(r.Context(), 1, metric.WithAttributes(attribute.String("key", "value")))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK!"))
		})
		svs.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
			upDown.Add(r.Context(), -1, metric.WithAttributes(attribute.String("key", "value")))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK!"))
		})
	}

	{ // histogram
		histogram, err := meter.Float64Histogram(
			"demo.request.duration",
			metric.WithDescription("Duration of histogram endpoint requests."),
			metric.WithUnit("s"),
		)
		log.Must(l, err, "failed to create histogram meter")

		svs.HandleFunc("/histogram", func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			time.Sleep(time.Second + time.Duration(rand.Intn(100))*time.Millisecond)

			histogram.Record(r.Context(), time.Since(start).Seconds(),
				metric.WithAttributes(attribute.String("key", "value")),
			)

			traceID := trace.SpanContextFromContext(r.Context())

			metricRequestLatency.(prometheus.ExemplarObserver).ObserveWithExemplar(
				time.Since(start).Seconds(), prometheus.Labels{"traceID": traceID.TraceID().String()},
			)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK!"))
		})
	}

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs,
			middleware.Telemetry(),
			middleware.Recover(),
		),
	)

	svr.Run()
}
