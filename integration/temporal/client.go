package keeltemporal

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.uber.org/zap"

	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
)

const defaultTracerName = "github.com/foomo/integration/temporal"

type (
	Options struct {
		Logger      *zap.Logger
		Namespace   *workflowservice.RegisterNamespaceRequest
		OtelEnabled bool
	}
	Option func(o *Options)
)

func WithOtelEnabled(v bool) Option {
	return func(o *Options) {
		o.OtelEnabled = env.GetBool("OTEL_ENABLED", v)
	}
}

func WithNamespace(v *workflowservice.RegisterNamespaceRequest) Option {
	return func(o *Options) {
		o.Namespace = v
	}
}

func DefaultOptions() Options {
	return Options{
		Logger:      log.Logger(),
		OtelEnabled: env.GetBool("OTEL_ENABLED", false),
		Namespace:   nil,
	}
}

func New(ctx context.Context, endpoint string, opts ...Option) (client.Client, error) {
	o := DefaultOptions()

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	clientOpts := client.Options{
		HostPort:  endpoint,
		Namespace: "default",
		Logger:    NewLogger(o.Logger),
	}

	// setup namespace
	if o.Namespace != nil {
		if nsc, err := client.NewNamespaceClient(clientOpts); err != nil {
			return nil, errors.Wrap(err, "failed to create temporal namespace client")
		} else if ns, err := nsc.Describe(ctx, o.Namespace.Namespace); err != nil {
			return nil, errors.Wrap(err, "failed to retrieve temporal namespace info")
		} else if ns.GetNamespaceInfo().State == enums.NAMESPACE_STATE_REGISTERED {
			o.Logger.Debug("temporal namespace already registered", log.FValue(o.Namespace))
		} else if err := nsc.Register(ctx, o.Namespace); err != nil {
			return nil, errors.Wrap(err, "failed to register temporal namespace")
		}
		clientOpts.Namespace = o.Namespace.Namespace
	}

	// setup otel
	if o.OtelEnabled {
		// bridgeTracer := otelopentracing.NewBridgeTracer()
		// bridgeTracer.SetOpenTelemetryTracer(telemetry.Tracer())
		tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
			Tracer:            telemetry.Tracer(),
			TextMapPropagator: otel.GetTextMapPropagator(),
			SpanContextKey:    nil,
			HeaderKey:         "",
			SpanStarter:       nil,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new opentracing interceptor")
		}
		clientOpts.Interceptors = append(clientOpts.Interceptors, tracingInterceptor)
		clientOpts.MetricsHandler = NewMetricsHandler(telemetry.MustMeter())
	}

	return client.NewClient(clientOpts)
}
