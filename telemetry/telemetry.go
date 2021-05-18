package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"

	"github.com/foomo/keel/env"
)

const (
	DefaultServiceName = "service"
	DefaultTracerName  = "github.com/foomo/keel/tracer"
)

var (
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
)

var (
	ServiceName = DefaultServiceName
	TracerName  = DefaultTracerName
)

func init() {
	if !env.GetBool("OTEL_ENABLED", false) {
		tracer = sdktrace.NewTracerProvider().Tracer(TracerName)
		return
	}

	ctx := context.Background()

	var opts []otlpgrpc.Option

	if env.GetBool("OTEL_EXPORTER_OTLP_INSECURE", false) {
		opts = append(opts, otlpgrpc.WithInsecure())
	}

	exp, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(opts...))
	if err != nil {
		panic(err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(env.Get("OTEL_SERVICE_NAME", ServiceName)),
		),
	)
	if err != nil {
		panic(err)
	}

	processor := sdktrace.NewBatchSpanProcessor(exp)
	provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(processor),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(provider)

	tracer = provider.Tracer(TracerName)
}

func Provider() *sdktrace.TracerProvider {
	return provider
}

func Tracer() trace.Tracer {
	return tracer
}

func Start(ctx context.Context, spanName string, opts ...trace.SpanOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}

func End(sp trace.Span, err error) {
	if err != nil {
		sp.RecordError(err)
		sp.SetStatus(codes.Error, err.Error())
	}
	sp.End()
}
