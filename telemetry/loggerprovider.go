package telemetry

import (
	"context"

	"github.com/foomo/keel/env"
	otelzap "github.com/foomo/keel/internal/otel/exporters/zap"
	otelzapbridge "go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// instrumentationName is the OTEL logger scope used by the zap bridge core.
const instrumentationName = "github.com/foomo/keel"

// LoggerProvider returns the global logger provider instance used throughout the application.
func LoggerProvider() log.LoggerProvider {
	return otel.GetLoggerProvider()
}

// NewNoopLoggerProvider returns a no-op [log.LoggerProvider].
func NewNoopLoggerProvider() log.LoggerProvider {
	return noop.NewLoggerProvider()
}

// NewZapBridgeCore returns a [zapcore.Core] that emits entries as OTEL log
// records through lp, for teeing into an existing logger.
func NewZapBridgeCore(lp log.LoggerProvider) zapcore.Core {
	return otelzapbridge.NewCore(instrumentationName, otelzapbridge.WithLoggerProvider(lp))
}

// NewZapLoggerProvider returns a logger provider that writes OTEL log
// records to logger and installs it as the global logger provider.
func NewZapLoggerProvider(ctx context.Context, logger *zap.Logger) (log.LoggerProvider, error) {
	exp, err := otelzap.New(logger)
	if err != nil {
		return nil, err
	}

	return newLoggerProvider(ctx, sdklog.NewSimpleProcessor(exp))
}

// NewStdOutLoggerProvider returns a logger provider that exports logs to
// stdout and installs it as the global logger provider. Output formatting is
// controlled by OTEL_EXPORTER_STDOUT_PRETTY_PRINT and
// OTEL_EXPORTER_STDOUT_TIMESTAMPS, both defaulting to true.
func NewStdOutLoggerProvider(ctx context.Context) (log.LoggerProvider, error) {
	var opts []stdoutlog.Option
	if env.GetBool("OTEL_EXPORTER_STDOUT_PRETTY_PRINT", true) {
		opts = append(opts, stdoutlog.WithPrettyPrint())
	}

	if !env.GetBool("OTEL_EXPORTER_STDOUT_TIMESTAMPS", true) {
		opts = append(opts, stdoutlog.WithoutTimestamps())
	}

	exp, err := stdoutlog.New(opts...)
	if err != nil {
		return nil, err
	}

	return newLoggerProvider(ctx, sdklog.NewSimpleProcessor(exp))
}

// NewOTLPHTTPLoggerProvider returns a logger provider that batches and
// exports logs over OTLP HTTP and installs it as the global logger provider.
// The exporter is configured from the OTEL_EXPORTER_OTLP_* environment.
func NewOTLPHTTPLoggerProvider(ctx context.Context) (log.LoggerProvider, error) {
	exp, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}

	return newLoggerProvider(ctx, sdklog.NewBatchProcessor(exp))
}

// NewOTLPGRPCLoggerProvider returns a logger provider that batches and
// exports logs over OTLP gRPC and installs it as the global logger provider.
// The exporter is configured from the OTEL_EXPORTER_OTLP_* environment.
func NewOTLPGRPCLoggerProvider(ctx context.Context) (log.LoggerProvider, error) {
	exp, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	return newLoggerProvider(ctx, sdklog.NewBatchProcessor(exp))
}

// NewOTLPGRCPLoggerProvider is a misspelled alias of
// [NewOTLPGRPCLoggerProvider].
//
// Deprecated: Use [NewOTLPGRPCLoggerProvider] instead.
func NewOTLPGRCPLoggerProvider(ctx context.Context) (log.LoggerProvider, error) {
	return NewOTLPGRPCLoggerProvider(ctx)
}

// newLoggerProvider creates an SDK logger provider with the default resource
// and processor p and installs it as the global logger provider.
func newLoggerProvider(ctx context.Context, p sdklog.Processor) (log.LoggerProvider, error) {
	resource, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(p),
	)
	otel.SetLoggerProvider(provider)

	return provider, nil
}
