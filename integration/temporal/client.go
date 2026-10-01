package keeltemporal

import (
	"context"

	goerrors "github.com/foomo/go/errors"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
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
)

type (
	// ClientOptions configures [NewClient].
	ClientOptions struct {
		// Logger receives the Temporal client logs.
		Logger *zap.Logger
		// Namespace is the namespace the client connects to.
		Namespace string
		// RegisterNamespace, when set, is registered if missing and updated
		// otherwise, and its namespace replaces Namespace.
		RegisterNamespace *workflowservice.RegisterNamespaceRequest
		// OtelEnabled reports whether tracing and metrics are enabled.
		OtelEnabled bool
		// ConnectionOptions are passed to the Temporal client unchanged.
		ConnectionOptions client.ConnectionOptions
	}
	// ClientOption modifies [ClientOptions].
	ClientOption func(o *ClientOptions)
)

// ClientWithOtelEnabled sets whether the client is instrumented with
// OpenTelemetry tracing and metrics.
func ClientWithOtelEnabled(v bool) ClientOption {
	return func(o *ClientOptions) {
		o.OtelEnabled = v
	}
}

// ClientWithNamespace sets the namespace the client connects to. Defaults to
// "default".
func ClientWithNamespace(v string) ClientOption {
	return func(o *ClientOptions) {
		o.Namespace = v
	}
}

// ClientWithRegisterNamespace sets a namespace that [NewClient] registers or
// updates before connecting to it.
func ClientWithRegisterNamespace(v *workflowservice.RegisterNamespaceRequest) ClientOption {
	return func(o *ClientOptions) {
		o.RegisterNamespace = v
	}
}

// ClientWithConnectionOptions sets the gRPC connection options of the client.
func ClientWithConnectionOptions(v client.ConnectionOptions) ClientOption {
	return func(o *ClientOptions) {
		o.ConnectionOptions = v
	}
}

// DefaultClientOptions returns the default [ClientOptions]: the keel logger,
// the "default" namespace and OpenTelemetry enabled according to the
// OTEL_TEMPORAL_ENABLED environment variable, falling back to OTEL_ENABLED
// and then false.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		Logger:            log.Logger(),
		Namespace:         "default",
		RegisterNamespace: nil,
		OtelEnabled:       env.GetBool("OTEL_TEMPORAL_ENABLED", env.GetBool("OTEL_ENABLED", false)),
		ConnectionOptions: client.ConnectionOptions{},
	}
}

// NewClient dials the Temporal server at endpoint (host:port) and returns a
// client configured by [DefaultClientOptions] and opts.
//
// If a namespace to register is set, it is registered when it does not exist
// and otherwise updated with the requested settings; an error is returned if
// the namespace is not in the registered state. If OpenTelemetry is enabled,
// a tracing interceptor and a metrics handler using the keel telemetry
// providers are installed.
func NewClient(ctx context.Context, endpoint string, opts ...ClientOption) (client.Client, error) {
	o := DefaultClientOptions()

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	clientOpts := client.Options{
		HostPort:          endpoint,
		Namespace:         o.Namespace,
		Logger:            NewLogger(o.Logger),
		ConnectionOptions: o.ConnectionOptions,
	}

	nsc, err := client.NewNamespaceClient(clientOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporal namespace client")
	}

	// setup namespace
	if o.RegisterNamespace != nil {
		ns, err := nsc.Describe(ctx, o.RegisterNamespace.Namespace)
		// Temporal's NamespaceClient.Describe returns *serviceerror.NamespaceNotFound on current
		// servers; older servers returned *serviceerror.NotFound. Both are treated as "missing".
		if goerrors.AsAnyType(err, &serviceerror.NotFound{}, &serviceerror.NamespaceNotFound{}) {
			if err := nsc.Register(ctx, o.RegisterNamespace); err != nil {
				return nil, errors.Wrap(err, "failed to register temporal namespace")
			}
		} else if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve temporal namespace info")
		}

		if ns.GetNamespaceInfo().GetState() != enums.NAMESPACE_STATE_REGISTERED { //nolint:nosnakecase
			return nil, errors.New("Could not register namespace due to existing state: " + ns.GetNamespaceInfo().GetState().String())
		}

		if err := nsc.Update(ctx, &workflowservice.UpdateNamespaceRequest{
			Namespace: o.RegisterNamespace.Namespace,
			UpdateInfo: &namespace.UpdateNamespaceInfo{
				Description: o.RegisterNamespace.Description,
				OwnerEmail:  o.RegisterNamespace.OwnerEmail,
				Data:        o.RegisterNamespace.Data,
				State:       ns.GetNamespaceInfo().GetState(),
			},
			Config: &namespace.NamespaceConfig{
				WorkflowExecutionRetentionTtl: o.RegisterNamespace.WorkflowExecutionRetentionPeriod,
				BadBinaries:                   ns.GetConfig().GetBadBinaries(),
				HistoryArchivalState:          o.RegisterNamespace.HistoryArchivalState,
				HistoryArchivalUri:            o.RegisterNamespace.HistoryArchivalUri,
				VisibilityArchivalState:       o.RegisterNamespace.VisibilityArchivalState,
				VisibilityArchivalUri:         o.RegisterNamespace.VisibilityArchivalUri,
			},
			ReplicationConfig: &replication.NamespaceReplicationConfig{
				ActiveClusterName: o.RegisterNamespace.ActiveClusterName,
				Clusters:          o.RegisterNamespace.Clusters,
				State:             ns.GetReplicationConfig().GetState(),
			},
			SecurityToken:    o.RegisterNamespace.SecurityToken,
			DeleteBadBinary:  "",
			PromoteNamespace: false,
		}); err != nil {
			return nil, errors.Wrap(err, "failed to update temporal namespace")
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
		clientOpts.MetricsHandler = opentelemetry.NewMetricsHandler(opentelemetry.MetricsHandlerOptions{
			Meter:   telemetry.Meter(),
			OnError: otel.Handle,
		})
	}

	return client.Dial(clientOpts)
}
