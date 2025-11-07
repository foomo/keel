package telemetry

import (
	"go.opentelemetry.io/otel/log"
)

func Logger(opts ...log.LoggerOption) log.Logger {
	return LoggerProvider().Logger(Name, opts...)
}
