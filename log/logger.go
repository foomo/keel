package log

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/env"
)

// config is the zap configuration of the last logger built by [NewLogger];
// atomicLevel is the level shared by all loggers built from it.
var (
	config      zap.Config
	atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

// init installs a logger configured from LOG_LEVEL (default "info") and
// LOG_FORMAT (default "json") as the zap global logger.
func init() {
	zap.ReplaceGlobals(NewLogger(
		env.Get("LOG_LEVEL", "info"),
		env.Get("LOG_FORMAT", "json"),
	))
}

// NewLogger builds a production zap logger with the given level (e.g. "info")
// and encoding ("json" or "console") and stores its configuration as the
// package configuration. It sets the shared [AtomicLevel] to level. Caller
// and stack traces are disabled unless level enables debug, overridable via
// LOG_DISABLE_CALLER and LOG_DISABLE_STACKTRACE. It panics if level is
// invalid or the logger cannot be built. The result is not installed as the
// global logger.
func NewLogger(level, encoding string) *zap.Logger {
	config = zap.NewProductionConfig()

	if value, err := zapcore.ParseLevel(level); err != nil {
		panic(err)
	} else {
		atomicLevel.SetLevel(value)
	}

	config.Encoding = encoding
	config.Level = atomicLevel
	config.EncoderConfig.TimeKey = "time"

	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	if encoding == "console" {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.CallerKey = AttributeKey(semconv.CodeFilePathKey)
	config.EncoderConfig.StacktraceKey = AttributeKey(semconv.CodeStacktraceKey)
	config.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	config.DisableCaller = env.GetBool("LOG_DISABLE_CALLER", !config.Level.Enabled(zap.DebugLevel))
	config.DisableStacktrace = env.GetBool("LOG_DISABLE_STACKTRACE", !config.Level.Enabled(zap.DebugLevel))

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

// Logger returns the zap global logger.
func Logger() *zap.Logger {
	return zap.L()
}

// AtomicLevel returns the level shared by loggers built with [NewLogger].
// Changing it adjusts their level at runtime.
func AtomicLevel() zap.AtomicLevel {
	return atomicLevel
}

// IsDisableCaller reports whether caller annotation is disabled.
func IsDisableCaller() bool {
	return config.DisableCaller
}

// IsDisableStacktrace reports whether automatic stack traces are disabled.
func IsDisableStacktrace() bool {
	return config.DisableStacktrace
}

// SetDisableCaller sets whether caller annotation is disabled and, if the
// value changed, rebuilds and installs the zap global logger. It returns an
// error if the logger cannot be built.
func SetDisableCaller(value bool) error {
	if value == config.DisableCaller {
		return nil
	}

	config.DisableCaller = value

	l, err := config.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(l)

	return nil
}

// SetDisableStacktrace sets whether automatic stack traces are disabled and,
// if the value changed, rebuilds and installs the zap global logger. It
// returns an error if the logger cannot be built.
func SetDisableStacktrace(value bool) error {
	if value == config.DisableStacktrace {
		return nil
	}

	config.DisableStacktrace = value

	l, err := config.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(l)

	return nil
}
