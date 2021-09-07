package telemetry

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	otelprometheus "go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	otelstdout "go.opentelemetry.io/otel/exporters/stdout"
	otelglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	otelcontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	otelsimple "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/env"
)

const (
	DefaultServiceName = "service"
	DefaultTracerName  = "github.com/foomo/keel/tracer"
)

var (
	tracer        trace.Tracer
	traceProvider trace.TracerProvider
	exporter      sdktrace.SpanExporter
	controller    *otelcontroller.Controller
	errorHandler  *ErrorHandler
)

var (
	ServiceName = DefaultServiceName
	TracerName  = DefaultTracerName
)

func init() {
	errorHandler = &ErrorHandler{}
	otel.SetErrorHandler(errorHandler)

	resource, err := otelresource.New(
		context.Background(),
		otelresource.WithAttributes(
			semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
		),
	)
	if err != nil {
		panic(err)
	}

	if env.GetBool("OTEL_ENABLED", false) {
		tp, c, e, err := newOtlp(resource)
		if err != nil {
			panic(err)
		}
		// TODO remove this once otlp <> tempo handles metrics
		_, err = NewPrometheus(resource)
		if err != nil {
			panic(err)
		}
		tracer = tp.Tracer(TracerName)
		traceProvider = tp
		controller = c
		exporter = e
	} else {
		tp, c, e, err := newStdOut(resource)
		if err != nil {
			panic(err)
		}
		tracer = tp.Tracer(TracerName)
		exporter = e
		traceProvider = tp
		controller = c
	}

	if env.GetBool("OTEL_METRICS_HOST_ENABLED", true) {
		if err := otelhost.Start(); err != nil {
			panic(err)
		}
	}
	if env.GetBool("OTEL_METRICS_RUNTIME_ENABLED", true) {
		if err := otelruntime.Start(); err != nil {
			panic(err)
		}
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func Controller() *otelcontroller.Controller {
	return controller
}

func Exporter() sdktrace.SpanExporter {
	return exporter
}

func TracerProvider() trace.TracerProvider {
	return traceProvider
}

func SetLogger(l *zap.Logger) {
	errorHandler.SetLogger(l)
}

func newStdOut(resource *otelresource.Resource) (*sdktrace.TracerProvider, *otelcontroller.Controller, *otelstdout.Exporter, error) {
	var exportOpts []otelstdout.Option
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		exportOpts = append(exportOpts, otelstdout.WithPrettyPrint())
	}
	if !env.GetBool("OTEL_EXPORTER_STDOUT_METRICS_ENABLED", false) {
		exportOpts = append(exportOpts, otelstdout.WithoutMetricExport())
	}
	if !env.GetBool("OTEL_EXPORTER_STDOUT_TRACE_ENABLED", false) {
		exportOpts = append(exportOpts, otelstdout.WithoutTraceExport())
	}

	exporter, err := otelstdout.NewExporter(exportOpts...)
	if err != nil {
		return nil, nil, nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	controller := otelcontroller.New(
		otelprocessor.New(
			otelsimple.NewWithInexpensiveDistribution(),
			exporter,
		),
		otelcontroller.WithExporter(exporter),
		otelcontroller.WithResource(resource),
	)
	err = controller.Start(context.Background())

	otel.SetTracerProvider(tracerProvider)
	otelglobal.SetMeterProvider(controller.MeterProvider())
	return tracerProvider, controller, exporter, err
}

func NewPrometheus(resource *otelresource.Resource) (*otelprometheus.Exporter, error) {
	return otelprometheus.InstallNewPipeline(
		otelprometheus.Config{
			Registerer: prometheus.DefaultRegisterer,
			Gatherer:   prometheus.DefaultGatherer,
		},
		otelcontroller.WithResource(resource),
	)
}

func newOtlp(resource *otelresource.Resource) (*sdktrace.TracerProvider, *otelcontroller.Controller, *otlp.Exporter, error) {
	var driverOpts []otlpgrpc.Option
	if env.GetBool("OTEL_EXPORTER_OTLP_INSECURE", false) {
		driverOpts = append(driverOpts, otlpgrpc.WithInsecure())
	}
	driver := otlpgrpc.NewDriver(driverOpts...)

	exporter, err := otlp.NewExporter(context.Background(), driver)
	if err != nil {
		return nil, nil, nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	controller := otelcontroller.New(
		otelprocessor.New(
			otelsimple.NewWithInexpensiveDistribution(),
			exporter,
		),
		otelcontroller.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)
	otelglobal.SetMeterProvider(controller.MeterProvider())

	if err = controller.Start(context.Background()); err != nil {
		return nil, nil, nil, err
	}

	return tracerProvider, controller, exporter, nil
}
