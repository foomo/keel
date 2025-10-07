package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/env"
)

var (
	config      zap.Config
	atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

func init() {
	zap.ReplaceGlobals(NewLogger(
		env.Get("LOG_LEVEL", "info"),
		env.Get("LOG_FORMAT", "json"),
	))
}

// NewLogger return a new logger instance
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
	config.EncoderConfig.CallerKey = "code_file_path"
	config.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(caller.File)
	}

	config.DisableCaller = env.GetBool("LOG_DISABLE_CALLER", config.Level.Enabled(zap.DebugLevel))
	config.DisableStacktrace = env.GetBool("LOG_DISABLE_STACKTRACE", !config.Level.Enabled(zap.DebugLevel))

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

// Logger return the logger instance
func Logger() *zap.Logger {
	return zap.L()
}

// AtomicLevel return the configured atomic level
func AtomicLevel() zap.AtomicLevel {
	return atomicLevel
}

// IsDisableCaller returns the configured disabled caller value
func IsDisableCaller() bool {
	return config.DisableCaller
}

// IsDisableStacktrace returns the configured disabled stacktrace value
func IsDisableStacktrace() bool {
	return config.DisableStacktrace
}

// SetDisableCaller sets the given value and re-configures the logger
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

// SetDisableStacktrace sets the given value and re-configures the logger
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

// Must logs a fatal error if given
func Must(l *zap.Logger, err error, msgAndArgs ...interface{}) {
	if err != nil {
		if l == nil {
			l = Logger()
		}

		var msg = "Must"
		if len(msgAndArgs) > 0 {
			msg, msgAndArgs = fmt.Sprintf("%v", msgAndArgs[0]), msgAndArgs[1:]
		}

		l.WithOptions(zap.AddCallerSkip(1)).Fatal(fmt.Sprintf(msg, msgAndArgs...), FError(err))
	}
}
