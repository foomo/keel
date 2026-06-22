package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/foomo/keel/env"
)

// Exporter env var values (OpenTelemetry spec). "console" maps to keel's StdOut providers.
const (
	exporterNone       = "none"
	exporterConsole    = "console"
	exporterOTLP       = "otlp"
	exporterPrometheus = "prometheus"
)

// OTLP protocol env var values (OpenTelemetry spec).
const (
	protocolGRPC = "grpc"
	protocolHTTP = "http/protobuf"
)

// otlpProtocol resolves the OTLP protocol for the given signal ("traces", "metrics", "logs").
// It honours the per-signal override OTEL_EXPORTER_OTLP_<SIGNAL>_PROTOCOL, then the global
// OTEL_EXPORTER_OTLP_PROTOCOL, defaulting to "http/protobuf" per the OpenTelemetry spec.
func otlpProtocol(signal string) string {
	key := "OTEL_EXPORTER_OTLP_" + strings.ToUpper(signal) + "_PROTOCOL"
	return env.Get(key, env.Get("OTEL_EXPORTER_OTLP_PROTOCOL", protocolHTTP))
}

// NewTraceProviderFromEnv selects a trace provider based on OTEL_TRACES_EXPORTER
// (none|console|otlp, default none) and OTEL_EXPORTER_OTLP_TRACES_PROTOCOL. It returns
// (nil, nil) when the exporter is none so the caller's noop fallback owns the default.
func NewTraceProviderFromEnv(ctx context.Context) (trace.TracerProvider, error) {
	switch exporter := env.Get("OTEL_TRACES_EXPORTER", exporterNone); exporter {
	case exporterNone:
		return nil, nil //nolint:nilnil // none disables the signal; nil lets the caller apply its noop fallback
	case exporterConsole:
		return NewStdOutTraceProvider(ctx)
	case exporterOTLP:
		switch protocol := otlpProtocol("traces"); protocol {
		case protocolGRPC:
			return NewOTLPGRPCTraceProvider(ctx)
		case protocolHTTP:
			return NewOTLPHTTPTraceProvider(ctx)
		default:
			return nil, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_TRACES_PROTOCOL: %q", protocol)
		}
	default:
		return nil, fmt.Errorf("unsupported OTEL_TRACES_EXPORTER: %q", exporter)
	}
}

// NewMeterProviderFromEnv selects a meter provider based on OTEL_METRICS_EXPORTER
// (none|console|otlp|prometheus, default none) and OTEL_EXPORTER_OTLP_METRICS_PROTOCOL.
// It returns (nil, nil) when the exporter is none so the caller's noop fallback owns the default.
func NewMeterProviderFromEnv(ctx context.Context) (metric.MeterProvider, error) {
	switch exporter := env.Get("OTEL_METRICS_EXPORTER", exporterNone); exporter {
	case exporterNone:
		return nil, nil //nolint:nilnil // none disables the signal; nil lets the caller apply its noop fallback
	case exporterConsole:
		return NewStdOutMeterProvider(ctx)
	case exporterPrometheus:
		return NewPrometheusMeterProvider(ctx)
	case exporterOTLP:
		switch protocol := otlpProtocol("metrics"); protocol {
		case protocolGRPC:
			return NewOTLPGRPCMeterProvider(ctx)
		case protocolHTTP:
			return NewOTLPHTTPMeterProvider(ctx)
		default:
			return nil, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_METRICS_PROTOCOL: %q", protocol)
		}
	default:
		return nil, fmt.Errorf("unsupported OTEL_METRICS_EXPORTER: %q", exporter)
	}
}

// NewLoggerProviderFromEnv selects a logger provider based on OTEL_LOGS_EXPORTER
// (none|console|otlp, default none) and OTEL_EXPORTER_OTLP_LOGS_PROTOCOL. It returns
// (nil, nil) when the exporter is none so the caller's noop fallback owns the default.
func NewLoggerProviderFromEnv(ctx context.Context) (log.LoggerProvider, error) {
	switch exporter := env.Get("OTEL_LOGS_EXPORTER", exporterNone); exporter {
	case exporterNone:
		return nil, nil //nolint:nilnil // none disables the signal; nil lets the caller apply its noop fallback
	case exporterConsole:
		return NewStdOutLoggerProvider(ctx)
	case exporterOTLP:
		switch protocol := otlpProtocol("logs"); protocol {
		case protocolGRPC:
			return NewOTLPGRCPLoggerProvider(ctx)
		case protocolHTTP:
			return NewOTLPHTTPLoggerProvider(ctx)
		default:
			return nil, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_LOGS_PROTOCOL: %q", protocol)
		}
	default:
		return nil, fmt.Errorf("unsupported OTEL_LOGS_EXPORTER: %q", exporter)
	}
}
