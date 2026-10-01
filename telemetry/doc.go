// Package telemetry provides OpenTelemetry tracing, metrics, logging and
// profiling helpers for keel services.
//
// # Wiring
//
// The usual entry point is [github.com/foomo/keel.WithTelemetry], which
// creates the trace, meter and logger providers from the environment through
// [NewTraceProviderFromEnv], [NewMeterProviderFromEnv] and
// [NewLoggerProviderFromEnv]. Every provider constructor in this package
// also installs its provider as the OpenTelemetry global, which [Tracer],
// [Meter] and [LoggerProvider] read.
//
// # Environment
//
// The exporters follow the OpenTelemetry environment variable specification:
//
//   - OTEL_TRACES_EXPORTER, OTEL_LOGS_EXPORTER: none (default), console or otlp.
//   - OTEL_METRICS_EXPORTER: none (default), console, otlp or prometheus.
//   - OTEL_EXPORTER_OTLP_PROTOCOL and OTEL_EXPORTER_OTLP_<SIGNAL>_PROTOCOL:
//     http/protobuf (default) or grpc.
//   - OTEL_EXPORTER_OTLP_* endpoint and header variables are read by the
//     OTLP exporters themselves.
//   - OTEL_SERVICE_NAME, OTEL_SERVICE_NAMESPACE and OTEL_RESOURCE_ATTRIBUTES
//     populate the resource, see [NewResource].
//
// keel specific variables:
//
//   - OTEL_TRACE_RATIO: parent based trace sampling ratio, defaults to 1.
//   - OTEL_METRICS_HOST_ENABLED, OTEL_METRICS_RUNTIME_ENABLED: start host and
//     Go runtime metric instrumentation.
//   - OTEL_EXPORTER_STDOUT_PRETTY_PRINT, OTEL_EXPORTER_STDOUT_TIMESTAMPS:
//     console exporter formatting, both default to true.
//   - OTEL_PROFILE_*: optional profile types of [NewProfiler].
//
// # Spans and logs
//
// [Ctx] wraps a [context.Context] in a [Context] that bundles span, log and
// profile helpers. Spans are named after the calling package and annotated
// with the caller's function, file and line:
//
//	func (s *Service) Do(ctx context.Context) (err error) {
//		tctx := telemetry.Ctx(ctx).StartSpan()
//		defer tctx.DeferEndSpan(&err)
//
//		tctx.LogInfo("doing work", attribute.String("id", s.id))
//
//		if err := s.work(tctx.Context()); err != nil {
//			tctx.RecordError(err)
//			return err
//		}
//
//		return nil
//	}
//
// Log messages are written through the global zap logger and carry the
// trace and span IDs of the context.
//
// # Metrics
//
// [NewIntCounter], [NewFloatHistogram] and their siblings create instruments
// from [Meter]. On error they report through [otel.Handle] and return a
// no-op instrument instead of nil.
package telemetry
