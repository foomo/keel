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

// Option func
type Option func(inst *Server)

// WithLogger option
func WithLogger(l *zap.Logger) Option {
	return func(inst *Server) {
		inst.l = l
	}
}

// WithLogFields option
func WithLogFields(fields ...zap.Field) Option {
	return func(inst *Server) {
		inst.l = inst.l.With(fields...)
	}
}

// WithConfig option
func WithConfig(c *viper.Viper) Option {
	return func(inst *Server) {
		inst.c = c
	}
}

// WithRemoteConfig option
func WithRemoteConfig(provider, endpoint, path string) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "config.remote.enabled", true)() {
			err := config.WithRemoteConfig(inst.c, provider, endpoint, path)
			log.Must(inst.l, err, "failed to add remote config")
		}
	}
}

// WithContext option
func WithContext(ctx context.Context) Option {
	return func(inst *Server) {
		inst.ctx = ctx
	}
}

// WithShutdownSignals option
func WithShutdownSignals(shutdownSignals ...os.Signal) Option {
	return func(inst *Server) {
		inst.shutdownSignals = shutdownSignals
	}
}

// WithGracefulPeriod option
func WithGracefulPeriod(gracefulPeriod time.Duration) Option {
	return func(inst *Server) {
		inst.gracefulPeriod = gracefulPeriod
	}
}

// WithHTTPZapService option with default value
func WithHTTPZapService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.zap.enabled", enabled)() {
			svs := service.NewDefaultHTTPZap(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPViperService option with default value
func WithHTTPViperService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.viper.enabled", enabled)() {
			svs := service.NewDefaultHTTPViper(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithStdOutTracer option with default value
func WithStdOutTracer(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewStdOutTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut trace provider")
		}
	}
}

// WithStdOutLogger option with default value
func WithStdOutLogger(enabled bool) Option {
	return func(inst *Server) {
		var err error

		_, err = telemetry.NewStdOutLoggerProvider(inst.ctx)
		log.Must(inst.l, err, "failed to create stdOut logger provider")
	}
}

// WithStdOutMeter option with default value
func WithStdOutMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewStdOutMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create stdOut meter provider")
		}
	}
}

// WithOTLPGRPCTracer option with default value
func WithOTLPGRPCTracer(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPGRPCTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp grpc trace provider")
		}
	}
}

// WithOTLPHTTPTracer option with default value
func WithOTLPHTTPTracer(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.traceProvider, err = telemetry.NewOTLPHTTPTraceProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create otlp http trace provider")
		}
	}
}

// WithOTLPGRPCMeter option with default value. Metrics are pushed via OTLP gRPC
// via a periodic reader, suiting setups that export metrics instead of exposing a
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

// WithOTLPHTTPMeter option with default value. Metrics are pushed via OTLP HTTP
// via a periodic reader, suiting setups that export metrics instead of exposing a
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

// WithPrometheusMeter option with default value
func WithPrometheusMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error

			inst.meterProvider, err = telemetry.NewPrometheusMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create prometheus meter provider")
		}
	}
}

// WithPushgatewayMeter option pushes Prometheus metrics to a Pushgateway on an
// interval and once more on graceful shutdown. An empty url disables it; the url
// falls back to the service.pushgateway.url config/env value. The push interval
// falls back to service.pushgateway.interval (default 15s). It sets up a Prometheus
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

// WithPyroscopeService option with default value
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

// WithHTTPPrometheusService option with default value
func WithHTTPPrometheusService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.prometheus.enabled", enabled)() {
			svs := service.NewDefaultHTTPPrometheus(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPPProfService option with default value
func WithHTTPPProfService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.pprof.enabled", enabled)() {
			svs := service.NewDefaultHTTPPProf(inst.Logger())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPHealthzService option with default value
func WithHTTPHealthzService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.healthz.enabled", enabled)() {
			svs := service.NewDefaultHTTPProbes(inst.Logger(), inst.probes())
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPReadmeService option with default value
func WithHTTPReadmeService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.readme.enabled", enabled)() {
			svs := service.NewDefaultHTTPReadme(inst.Logger(), inst.readmers)
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithInitService option with default value
func WithInitService(service Service) Option {
	return func(inst *Server) {
		if service == nil || slices.Contains(inst.initServices, service) {
			return
		}

		inst.initServices = append(inst.initServices, service)
	}
}
