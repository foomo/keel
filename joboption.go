package keel

import (
	"context"
	"os"
	"time"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// JobOption configures a [Job] in [NewJob].
type JobOption func(inst *Job)

// JobWithLogger sets the job logger. Defaults to [log.Logger].
func JobWithLogger(l *zap.Logger) JobOption {
	return func(inst *Job) {
		inst.l = l
	}
}

// JobWithLogFields adds fields to the job logger.
func JobWithLogFields(fields ...zap.Field) JobOption {
	return func(inst *Job) {
		inst.l = inst.l.With(fields...)
	}
}

// JobWithName overrides the default job name (OTEL_SERVICE_NAME, falling back
// to [telemetry.DefaultServiceName]).
func JobWithName(name string) JobOption {
	return func(inst *Job) {
		inst.name = name
	}
}

// JobWithConfig sets the job configuration. Defaults to [config.Config].
func JobWithConfig(c *viper.Viper) JobOption {
	return func(inst *Job) {
		inst.c = c
	}
}

// JobWithParallel runs the job steps concurrently instead of sequentially.
// limit caps the number of steps running at once; limit <= 0 means unbounded. The
// first failing step cancels the rest (fail-fast) and [Job.RunE] returns the joined error.
func JobWithParallel(limit int) JobOption {
	return func(inst *Job) {
		inst.parallel = true
		inst.parallelLimit = limit
	}
}

// JobWithContext sets the parent of the job context. Defaults to
// [context.Background].
func JobWithContext(ctx context.Context) JobOption {
	return func(inst *Job) {
		inst.ctx = ctx
	}
}

// JobWithShutdownSignals sets the signals that interrupt the job. Defaults to
// SIGINT and SIGTERM.
func JobWithShutdownSignals(shutdownSignals ...os.Signal) JobOption {
	return func(inst *Job) {
		inst.shutdownSignals = shutdownSignals
	}
}

// JobWithGracefulPeriod sets the budget for flushing telemetry and closing
// resources after the job completes or is interrupted. Defaults to
// KEEL_GRACEFUL_PERIOD seconds, or 30 seconds.
func JobWithGracefulPeriod(gracefulPeriod time.Duration) JobOption {
	return func(inst *Job) {
		inst.gracefulPeriod = gracefulPeriod
	}
}

// JobWithTimeout sets an in-process deadline for the whole job run.
// A zero duration (default) disables the in-process deadline; Kubernetes
// activeDeadlineSeconds still applies via SIGTERM handling.
func JobWithTimeout(timeout time.Duration) JobOption {
	return func(inst *Job) {
		inst.timeout = timeout
	}
}

// JobWithCloser registers closer to be called during job finalization. See
// [Job.AddCloser].
func JobWithCloser(closer any) JobOption {
	return func(inst *Job) {
		inst.AddCloser(closer)
	}
}

// JobWithTelemetry wires the OpenTelemetry trace, metric and logger providers
// from the standard OTEL environment variables:
//
//	OTEL_TRACES_EXPORTER   none(default) | console | otlp
//	OTEL_METRICS_EXPORTER  none(default) | console | otlp | prometheus
//	OTEL_LOGS_EXPORTER     none(default) | console | otlp
//	OTEL_EXPORTER_OTLP_PROTOCOL  grpc | http/protobuf(default)
//	  (per-signal override: OTEL_EXPORTER_OTLP_{TRACES,METRICS,LOGS}_PROTOCOL)
//
// A signal set to none leaves its provider unset, so the job falls back to a no-op
// provider. Call this before [JobWithPushgatewayMeter] so its nil meter-provider guard
// still fires when OTEL_METRICS_EXPORTER is none.
func JobWithTelemetry() JobOption {
	return func(inst *Job) {
		traceProvider, err := telemetry.NewTraceProviderFromEnv(inst.ctx)
		log.Must(inst.l, err, "failed to create trace provider")

		if traceProvider != nil {
			inst.traceProvider = traceProvider
		}

		meterProvider, err := telemetry.NewMeterProviderFromEnv(inst.ctx)
		log.Must(inst.l, err, "failed to create meter provider")

		if meterProvider != nil {
			inst.meterProvider = meterProvider
		}

		loggerProvider, err := telemetry.NewLoggerProviderFromEnv(inst.ctx)
		log.Must(inst.l, err, "failed to create logger provider")

		if loggerProvider != nil {
			inst.loggerProvider = loggerProvider
		}
	}
}

// JobWithStdOutTracer sets a trace provider that writes to stdout. enabled is
// the default for the otel.enabled config key.
func JobWithStdOutTracer(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewStdOutTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut trace provider")
		}
	}
}

// JobWithStdOutMeter sets a meter provider that writes to stdout. enabled is
// the default for the otel.enabled config key.
func JobWithStdOutMeter(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewStdOutMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut meter provider")
		}
	}
}

// JobWithOTLPGRPCTracer sets an OTLP gRPC trace provider. enabled is the
// default for the otel.enabled config key.
func JobWithOTLPGRPCTracer(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPGRPCTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc trace provider")
		}
	}
}

// JobWithOTLPHTTPTracer sets an OTLP HTTP trace provider. enabled is the
// default for the otel.enabled config key.
func JobWithOTLPHTTPTracer(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPHTTPTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http trace provider")
		}
	}
}

// JobWithOTLPGRPCMeter sets an OTLP gRPC meter provider. enabled is the default
// for the otel.enabled config key. Metrics are pushed via OTLP gRPC
// and flushed on job exit, suiting jobs that finish before a Prometheus scrape.
func JobWithOTLPGRPCMeter(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewOTLPGRPCMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc meter provider")
		}
	}
}

// JobWithOTLPHTTPMeter sets an OTLP HTTP meter provider. enabled is the default
// for the otel.enabled config key. Metrics are pushed via OTLP HTTP
// and flushed on job exit, suiting jobs that finish before a Prometheus scrape.
func JobWithOTLPHTTPMeter(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewOTLPHTTPMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http meter provider")
		}
	}
}

// JobWithOTLPHTTPLogger sets an OTLP HTTP logger provider. enabled is the
// default for the otel.enabled config key.
func JobWithOTLPHTTPLogger(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewOTLPHTTPLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http logger provider")
		}
	}
}

// JobWithOTLPGRCPLogger sets an OTLP gRPC logger provider. enabled is the
// default for the otel.enabled config key.
func JobWithOTLPGRCPLogger(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewOTLPGRPCLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc logger provider")
		}
	}
}

// JobWithPushgatewayMeter pushes Prometheus metrics to a Pushgateway on job
// exit, grouped by the job name. url is the default for the
// service.pushgateway.url config key (env SERVICE_PUSHGATEWAY_URL); an empty
// result disables it. It sets up a Prometheus meter provider (unless one is
// already configured) so OTEL metrics are included in the push.
func JobWithPushgatewayMeter(url string) JobOption {
	return func(inst *Job) {
		url = config.GetString(inst.Config(), "service.pushgateway.url", url)()
		if url == "" {
			return
		}

		if inst.meterProvider == nil {
			var err error

			inst.meterProvider, err = telemetry.NewPrometheusMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create prometheus meter provider")
		}

		inst.pushers = append(inst.pushers, func(ctx context.Context) error {
			return telemetry.PushToGateway(ctx, url, inst.name)
		})
	}
}
