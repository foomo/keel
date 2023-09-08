package keel

import (
	"context"
	"os"
	"time"

	"github.com/foomo/keel/service"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
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

// WithShutdownTimeout option
func WithShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(inst *Server) {
		inst.shutdownTimeout = shutdownTimeout
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
			log.Must(inst.l, err, "failed to create std out trace provider")
		}
	}
}

// WithStdOutMeter option with default value
func WithStdOutMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error
			inst.meterProvider, err = telemetry.NewStdOutMeterProvider(inst.ctx)
			log.Must(inst.l, err, "failed to create std out meter provider")
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

// WithPrometheusMeter option with default value
func WithPrometheusMeter(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "otel.enabled", enabled)() {
			var err error
			inst.meterProvider, err = telemetry.NewPrometheusMeterProvider()
			log.Must(inst.l, err, "failed to create prometheus meter provider")
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
			svs := service.NewDefaultHTTPProbes(inst.Logger(), inst.probes)
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}

// WithHTTPReadmeService option with default value
func WithHTTPReadmeService(enabled bool) Option {
	return func(inst *Server) {
		if config.GetBool(inst.Config(), "service.readme.enabled", enabled)() {
			svs := service.NewDefaultHTTPReadme(inst.Logger(), &inst.readmers)
			inst.initServices = append(inst.initServices, svs)
			inst.AddAlwaysHealthzers(svs)
		}
	}
}
