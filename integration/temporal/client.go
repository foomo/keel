package keeltemporal

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.uber.org/zap"

	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
)

type (
	ClientOptions struct {
		Logger            *zap.Logger
		Namespace         string
		RegisterNamespace *workflowservice.RegisterNamespaceRequest
		OtelEnabled       bool
	}
	ClientOption func(o *ClientOptions)
)

func ClientWithOtelEnabled(v bool) ClientOption {
	return func(o *ClientOptions) {
		o.OtelEnabled = v
	}
}

func ClientWithNamespace(v string) ClientOption {
	return func(o *ClientOptions) {
		o.Namespace = v
	}
}

func ClientWithRegisterNamespace(v *workflowservice.RegisterNamespaceRequest) ClientOption {
	return func(o *ClientOptions) {
		o.RegisterNamespace = v
	}
}

func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		Logger:            log.Logger(),
		Namespace:         "default",
		RegisterNamespace: nil,
		OtelEnabled:       env.GetBool("OTEL_TEMPORAL_ENABLED", env.GetBool("OTEL_ENABLED", false)),
	}
}

func NewClient(ctx context.Context, endpoint string, opts ...ClientOption) (client.Client, error) {
	o := DefaultClientOptions()

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	clientOpts := client.Options{
		HostPort:  endpoint,
		Namespace: o.Namespace,
		Logger:    NewLogger(o.Logger),
	}

	nsc, err := client.NewNamespaceClient(clientOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporal namespace client")
	}

	// setup namespace
	if o.RegisterNamespace != nil {
		var notFoundErr *serviceerror.NotFound
		if ns, err := nsc.Describe(ctx, o.RegisterNamespace.Namespace); errors.As(err, &notFoundErr) {
			if err := nsc.Register(ctx, o.RegisterNamespace); err != nil {
				return nil, errors.Wrap(err, "failed to register temporal namespace")
			}
		} else if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve temporal namespace info")
		} else if nsInfo := ns.GetNamespaceInfo(); nsInfo.State != enums.NAMESPACE_STATE_REGISTERED { //nolint:nosnakecase
			return nil, errors.New("Could not register namespace due to existing state: " + nsInfo.State.String())
		} else if err := nsc.Update(ctx, &workflowservice.UpdateNamespaceRequest{
			Namespace: o.RegisterNamespace.Namespace,
			UpdateInfo: &namespace.UpdateNamespaceInfo{
				Description: o.RegisterNamespace.Description,
				OwnerEmail:  o.RegisterNamespace.OwnerEmail,
				Data:        o.RegisterNamespace.Data,
				State:       nsInfo.State,
			},
			Config: &namespace.NamespaceConfig{
				WorkflowExecutionRetentionTtl: o.RegisterNamespace.WorkflowExecutionRetentionPeriod,
				BadBinaries:                   ns.Config.BadBinaries,
				HistoryArchivalState:          o.RegisterNamespace.HistoryArchivalState,
				HistoryArchivalUri:            o.RegisterNamespace.HistoryArchivalUri,
				VisibilityArchivalState:       o.RegisterNamespace.VisibilityArchivalState,
				VisibilityArchivalUri:         o.RegisterNamespace.VisibilityArchivalUri,
			},
			ReplicationConfig: &replication.NamespaceReplicationConfig{
				ActiveClusterName: o.RegisterNamespace.ActiveClusterName,
				Clusters:          o.RegisterNamespace.Clusters,
				State:             ns.ReplicationConfig.State,
			},
			SecurityToken:    o.RegisterNamespace.SecurityToken,
			DeleteBadBinary:  "",
			PromoteNamespace: false,
		}); err != nil {
			return nil, errors.Wrap(err, "failed to register temporal namespace")
		}
		clientOpts.Namespace = o.RegisterNamespace.Namespace
	}

	// setup otel
	if o.OtelEnabled {
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
		clientOpts.MetricsHandler = NewMetricsHandler(telemetry.Meter())
	}

	return client.Dial(clientOpts)
}
