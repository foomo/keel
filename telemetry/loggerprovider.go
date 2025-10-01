package telemetry

import (
	"context"

	"github.com/foomo/keel/env"
	otelzap "github.com/foomo/keel/internal/otel/exporters/zap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
)

// LoggerProvider returns the global logger provider instance used throughout the application.
func LoggerProvider() log.LoggerProvider {
	return global.GetLoggerProvider()
}

// NewZapLoggerProvider creates a new log.LoggerProvider using a Zap logger for structured logging with OpenTelemetry.
func NewZapLoggerProvider(ctx context.Context, logger *zap.Logger) (log.LoggerProvider, error) {
	exp, err := otelzap.New(logger)
	if err != nil {
		return nil, err
	}
	return newLoggerProvider(ctx, sdklog.NewSimpleProcessor(exp))
}

// NewStdOutLoggerProvider creates a logger provider that exports logs to standard output with configurable options.
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

// NewOTLPHTTPLoggerProvider creates a new OTLP HTTP LoggerProvider with a batch processor and default resource.
func NewOTLPHTTPLoggerProvider(ctx context.Context) (log.LoggerProvider, error) {
	exp, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}

	return newLoggerProvider(ctx, sdklog.NewBatchProcessor(exp))
}

// NewOTLPGRCPLoggerProvider creates a new OTLP gRPC-based logger provider using the provided context.
func NewOTLPGRCPLoggerProvider(ctx context.Context) (log.LoggerProvider, error) {
	exp, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	return newLoggerProvider(ctx, sdklog.NewBatchProcessor(exp))
}

func newLoggerProvider(ctx context.Context, p sdklog.Processor) (log.LoggerProvider, error) {
	resource, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(p),
	)
	global.SetLoggerProvider(provider)
	return provider, nil
}
