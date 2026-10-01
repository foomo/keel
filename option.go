package keel

import (
	"context"
	"os"
	"slices"
	"time"

	"github.com/foomo/keel/service"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
)

// Option configures a [Server] in [NewServer].
type Option func(inst *Server)

// WithLogger sets the server logger. Defaults to [log.Logger].
func WithLogger(l *zap.Logger) Option {
	return func(inst *Server) {
		inst.l = l
	}
}

// WithLogFields adds fields to the server logger.
func WithLogFields(fields ...zap.Field) Option {
	return func(inst *Server) {
		inst.l = inst.l.With(fields...)
	}
}

// WithConfig sets the server configuration. Defaults to [config.Config].
func WithConfig(c *viper.Viper) Option {
	return func(inst *Server) {
		inst.c = c
	}
}

// WithRemoteConfig adds a remote configuration source using
// [config.WithRemoteConfig] unless config.remote.enabled is false. Failing to
// add the source is fatal.
func WithRemoteConfig(provider, endpoint, path string) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "config.remote.enabled", true)() {
			err := config.WithRemoteConfig(inst.c, provider, endpoint, path)
			log.Must(inst.l, err, "failed to add remote config")
		}
	}
}

// WithContext sets the parent of the server context. Defaults to
// [context.Background].
func WithContext(ctx context.Context) Option {
	return func(inst *Server) {
		inst.ctx = ctx
	}
}

// WithShutdownSignals sets the signals that trigger graceful shutdown. Defaults
// to SIGINT and SIGTERM.
func WithShutdownSignals(shutdownSignals ...os.Signal) Option {
	return func(inst *Server) {
		inst.shutdownSignals = shutdownSignals
	}
}

// WithGracefulPeriod sets the time budget for closing resources on graceful
// shutdown. It should match the pod's terminationGracePeriodSeconds. Defaults
// to KEEL_GRACEFUL_PERIOD seconds, or 30 seconds.
func WithGracefulPeriod(gracefulPeriod time.Duration) Option {
	return func(inst *Server) {
		inst.gracefulPeriod = gracefulPeriod
	}
}

// WithHTTPZapService registers the log level service from
// [service.NewDefaultHTTPZap] as an init service. enabled is the default for the
// service.zap.enabled config key.
func WithHTTPZapService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.zap.enabled", enabled)() {
			svs := service.NewDefaultHTTPZap(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPViperService registers the configuration service from
// [service.NewDefaultHTTPViper] as an init service. enabled is the default for
// the service.viper.enabled config key.
func WithHTTPViperService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.viper.enabled", enabled)() {
			svs := service.NewDefaultHTTPViper(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithTelemetry wires the OpenTelemetry trace, metric and logger providers
// from the standard OTEL environment variables:
//
//	OTEL_TRACES_EXPORTER   none(default) | console | otlp
//	OTEL_METRICS_EXPORTER  none(default) | console | otlp | prometheus
//	OTEL_LOGS_EXPORTER     none(default) | console | otlp
//	OTEL_EXPORTER_OTLP_PROTOCOL  grpc | http/protobuf(default)
//	  (per-signal override: OTEL_EXPORTER_OTLP_{TRACES,METRICS,LOGS}_PROTOCOL)
//
// A signal set to none leaves its provider unset, so the server falls back to a
// no-op provider. Call this before [WithPushgatewayMeter] so its nil meter-provider
// guard still fires when OTEL_METRICS_EXPORTER is none.
func WithTelemetry() Option {
	return func(inst *Server) {
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

// WithStdOutTracer sets a trace provider that writes to stdout. enabled is the
// default for the otel.enabled config key.
func WithStdOutTracer(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewStdOutTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut trace provider")
		}
	}
}

// WithStdOutLogger sets a logger provider that writes to stdout. enabled is the
// default for the otel.enabled config key.
func WithStdOutLogger(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewStdOutLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut logger provider")
		}
	}
}

// WithOTLPHTTPLogger sets an OTLP HTTP logger provider. enabled is the default
// for the otel.enabled config key.
func WithOTLPHTTPLogger(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewOTLPHTTPLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http logger provider")
		}
	}
}

// WithOTLPGRCPLogger sets an OTLP gRPC logger provider. enabled is the default
// for the otel.enabled config key.
func WithOTLPGRCPLogger(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.loggerProvider, err = telemetry.NewOTLPGRPCLoggerProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc logger provider")
		}
	}
}

// WithStdOutMeter sets a meter provider that writes to stdout. enabled is the
// default for the otel.enabled config key.
func WithStdOutMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewStdOutMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut meter provider")
		}
	}
}

// WithOTLPGRPCTracer sets an OTLP gRPC trace provider. enabled is the default
// for the otel.enabled config key.
func WithOTLPGRPCTracer(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPGRPCTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc trace provider")
		}
	}
}

// WithOTLPHTTPTracer sets an OTLP HTTP trace provider. enabled is the default
// for the otel.enabled config key.
func WithOTLPHTTPTracer(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPHTTPTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http trace provider")
		}
	}
}

// WithOTLPGRPCMeter sets an OTLP gRPC meter provider. enabled is the default for
// the otel.enabled config key. Metrics are pushed by a periodic reader, suiting setups that export metrics instead of exposing a
// Prometheus scrape endpoint.
func WithOTLPGRPCMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewOTLPGRPCMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc meter provider")
		}
	}
}

// WithOTLPHTTPMeter sets an OTLP HTTP meter provider. enabled is the default for
// the otel.enabled config key. Metrics are pushed by a periodic reader, suiting setups that export metrics instead of exposing a
// Prometheus scrape endpoint.
func WithOTLPHTTPMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewOTLPHTTPMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http meter provider")
		}
	}
}

// WithPrometheusMeter sets a Prometheus meter provider. enabled is the default
// for the otel.enabled config key.
func WithPrometheusMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewPrometheusMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create prometheus meter provider")
		}
	}
}

// WithPushgatewayMeter registers an init service that pushes Prometheus metrics
// to a Pushgateway on an interval and once more on graceful shutdown. url is the
// default for the service.pushgateway.url config key; an empty result disables
// it. The push interval is read from service.pushgateway.interval (default 15s). It sets up a Prometheus
// meter provider (unless one is already configured) so OTEL metrics are included in
// the push.
func WithPushgatewayMeter(url string) Option {
	return func(inst *Server) {
		url = config.GetString(inst.Config(), "service.pushgateway.url", url)()
		if url == "" {
			return
		}

		interval := config.GetDuration(inst.Config(), "service.pushgateway.interval", 15*time.Second)()

		if inst.meterProvider == nil {
			var err error

			inst.meterProvider, err = telemetry.NewPrometheusMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create prometheus meter provider")
		}

		name := env.Get("OTEL_SERVICE_NAME", telemetry.DefaultServiceName)

		svs := service.NewGoRoutine(inst.Logger(), "pushgateway", func(ctx context.Context, l *zap.Logger) error {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := telemetry.PushToGateway(ctx, url, name); err != nil {
						log.WithError(l, err).Warn("failed to push to pushgateway")
					}
				case <-ctx.Done():
					// final push detached from cancellation so it runs during shutdown
					pushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), interval)
					defer cancel()

					if err := telemetry.PushToGateway(pushCtx, url, name); err != nil {
						log.WithError(l, err).Warn("failed to push to pushgateway on shutdown")
					}

					l.Info("stopping pushgateway")

					return nil
				}
			}
		})
		inst.initServices = append(inst.initServices, svs)
		inst.AddAlwaysHealthzers(svs)
	}
}

// WithPyroscopeService registers an init service running the Pyroscope profiler
// from [telemetry.NewProfiler]. enabled is the default for the otel.enabled
// config key.
func WithPyroscopeService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			svs := service.NewGoRoutine(inst.Logger(), "pyroscope", func(ctx context.Context, l *zap.Logger) error {
				p, err := telemetry.NewProfiler(ctx)
				if err != nil {
					return err
				}

				<-ctx.Done()
				p.Flush(true)
				l.Info("stopping pyroscope")

				return p.Stop()
			})
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPPrometheusService registers the metrics service from
// [service.NewDefaultHTTPPrometheus] as an init service. enabled is the default
// for the service.prometheus.enabled config key.
func WithHTTPPrometheusService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.prometheus.enabled", enabled)() {
			svs := service.NewDefaultHTTPPrometheus(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPPProfService registers the pprof service from
// [service.NewDefaultHTTPPProf] as an init service. enabled is the default for
// the service.pprof.enabled config key.
func WithHTTPPProfService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.pprof.enabled", enabled)() {
			svs := service.NewDefaultHTTPPProf(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPHealthzService registers the health probe service from
// [service.NewDefaultHTTPProbes] as an init service. enabled is the default for
// the service.healthz.enabled config key.
func WithHTTPHealthzService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.healthz.enabled", enabled)() {
			svs := service.NewDefaultHTTPProbes(inst.Logger(), inst.probes())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPReadmeService registers the readme service from
// [service.NewDefaultHTTPReadme] as an init service. enabled is the default for
// the service.readme.enabled config key.
//
// Deprecated: Readme support will be removed in a future release.
func WithHTTPReadmeService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.readme.enabled", enabled)() {
			svs := service.NewDefaultHTTPReadme(inst.Logger(), inst.readmers)
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithInitService registers service as an init service, started in [NewServer].
// Nil and already registered services are ignored.
func WithInitService(service Service) Option {
	return func(inst *Server) {
		if service == nil || slices.Contains(inst.initServices, service) {
			return
		}

		inst.initServices = append(inst.initServices, service)
	}
}
