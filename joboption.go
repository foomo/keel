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

// JobOption func
type JobOption func(inst *Job)

// JobWithLogger option
func JobWithLogger(l *zap.Logger) JobOption {
	return func(inst *Job) {
		inst.l = l
	}
}

// JobWithLogFields option
func JobWithLogFields(fields ...zap.Field) JobOption {
	return func(inst *Job) {
		inst.l = inst.l.With(fields...)
	}
}

// JobWithName option overrides the default job name (OTEL_SERVICE_NAME, falling
// back to telemetry.DefaultServiceName).
func JobWithName(name string) JobOption {
	return func(inst *Job) {
		inst.name = name
	}
}

// JobWithConfig option
func JobWithConfig(c *viper.Viper) JobOption {
	return func(inst *Job) {
		inst.c = c
	}
}

// JobWithParallel option runs the job steps concurrently instead of sequentially.
// limit caps the number of steps running at once; limit <= 0 means unbounded. The
// first failing step cancels the rest (fail-fast) and RunE returns the joined error.
func JobWithParallel(limit int) JobOption {
	return func(inst *Job) {
		inst.parallel = true
		inst.parallelLimit = limit
	}
}

// JobWithContext option
func JobWithContext(ctx context.Context) JobOption {
	return func(inst *Job) {
		inst.ctx = ctx
	}
}

// JobWithShutdownSignals option
func JobWithShutdownSignals(shutdownSignals ...os.Signal) JobOption {
	return func(inst *Job) {
		inst.shutdownSignals = shutdownSignals
	}
}

// JobWithGracefulPeriod option sets the budget for flushing telemetry and closing
// resources after the job completes or is interrupted.
func JobWithGracefulPeriod(gracefulPeriod time.Duration) JobOption {
	return func(inst *Job) {
		inst.gracefulPeriod = gracefulPeriod
	}
}

// JobWithTimeout option sets an in-process deadline for the whole job run.
// A zero duration (default) disables the in-process deadline; Kubernetes
// activeDeadlineSeconds still applies via SIGTERM handling.
func JobWithTimeout(timeout time.Duration) JobOption {
	return func(inst *Job) {
		inst.timeout = timeout
	}
}

// JobWithCloser option registers a closer to be called during job finalization.
func JobWithCloser(closer any) JobOption {
	return func(inst *Job) {
		inst.AddCloser(closer)
	}
}

// JobWithTelemetry option wires the OpenTelemetry trace, metric and logger providers
// from the standard OTEL environment variables:
//
//	OTEL_TRACES_EXPORTER   none(default) | console | otlp
//	OTEL_METRICS_EXPORTER  none(default) | console | otlp | prometheus
//	OTEL_LOGS_EXPORTER     none(default) | console | otlp
//	OTEL_EXPORTER_OTLP_PROTOCOL  grpc | http/protobuf(default)
//	  (per-signal override: OTEL_EXPORTER_OTLP_{TRACES,METRICS,LOGS}_PROTOCOL)
//
// A signal set to none leaves its provider unset, so the job falls back to a no-op
// provider. Call this before JobWithPushgatewayMeter so its nil meter-provider guard
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

// JobWithStdOutTracer option with default value.
func JobWithStdOutTracer(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewStdOutTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut trace provider")
		}
	}
}

// JobWithStdOutMeter option with default value.
func JobWithStdOutMeter(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewStdOutMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut meter provider")
		}
	}
}

// JobWithOTLPGRPCTracer option with default value.
func JobWithOTLPGRPCTracer(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPGRPCTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc trace provider")
		}
	}
}

// JobWithOTLPHTTPTracer option with default value.
func JobWithOTLPHTTPTracer(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPHTTPTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http trace provider")
		}
	}
}

// JobWithOTLPGRPCMeter option with default value. Metrics are pushed via OTLP gRPC
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

// JobWithOTLPHTTPMeter option with default value. Metrics are pushed via OTLP HTTP
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

// JobWithOTLPHTTPLogger option with default value
func JobWithOTLPHTTPLogger(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewOTLPHTTPLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http logger provider")
		}
	}
}

// JobWithOTLPGRCPLogger option with default value
func JobWithOTLPGRCPLogger(enabled bool) JobOption {
	return func(inst *Job) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewOTLPGRCPLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc logger provider")
		}
	}
}

// JobWithPushgatewayMeter option pushes Prometheus metrics to a Pushgateway on job
// exit. An empty url disables it; the url falls back to the KEEL_PUSHGATEWAY_URL
// config/env value. It sets up a Prometheus meter provider so OTEL metrics are
// included in the push.
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
