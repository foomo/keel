package telemetry

import (
	"context"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	otelglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/nonrecording"
	otelcontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelaggregation "go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	otelselector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/foomo/keel/env"
)

func Meter() metric.Meter {
	return otelglobal.Meter("")
}

func NewNoopMeterProvider() (metric.MeterProvider, error) {
	controller := nonrecording.NewNoopMeterProvider()
	otelglobal.SetMeterProvider(controller)
	return controller, nil
}

func NewStdOutMeterProvider(ctx context.Context, opts ...stdoutmetric.Option) (metric.MeterProvider, error) {
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		opts = append(opts, stdoutmetric.WithPrettyPrint())
	}

	exporter, err := stdoutmetric.New(opts...)
	if err != nil {
		log.Fatalf("creating stdoutmetric exporter: %v", err)
	}

	resource := otelresource.NewSchemaless(
		semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
	)

	controller := otelcontroller.New(
		otelprocessor.NewFactory(
			otelselector.NewWithInexpensiveDistribution(),
			exporter,
		),
		otelcontroller.WithExporter(exporter),
		otelcontroller.WithResource(resource),
	)

	if err = controller.Start(ctx); err != nil {
		return nil, err
	}

	otelglobal.SetMeterProvider(controller)
	return controller, nil
}

func NewPrometheusMeterProvider() (metric.MeterProvider, error) {
	config := otelprometheus.Config{
		Registerer: prometheus.DefaultRegisterer,
		Gatherer:   prometheus.DefaultGatherer,
	}

	resource := otelresource.NewSchemaless(
		semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
	)

	controller := otelcontroller.New(
		otelprocessor.NewFactory(
			otelselector.NewWithHistogramDistribution(),
			otelaggregation.CumulativeTemporalitySelector(),
			otelprocessor.WithMemory(true),
		),
		otelcontroller.WithResource(resource),
	)

	_, err := otelprometheus.New(config, controller)
	if err != nil {
		return nil, err
	}

	otelglobal.SetMeterProvider(controller)
	return controller, nil
}
