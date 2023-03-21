package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/env"
)

const (
	ModeDev  = "dev"
	ModeProd = "prod"
)

var (
	config      zap.Config
	atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

func init() {
	var level string
	switch env.Get("LOG_MODE", ModeProd) {
	case ModeDev:
		config = NewDevelopmentConfig()
		level = env.Get("LOG_LEVEL", "debug")
	default:
		config = NewProductionConfig()
		level = env.Get("LOG_LEVEL", "info")
	}
	config.Level = atomicLevel
	config.EncoderConfig.TimeKey = "time"
	config.DisableCaller = env.GetBool("LOG_DISABLE_STACKTRACE", true)
	config.DisableStacktrace = env.GetBool("LOG_DISABLE_CALLER", true)

	if value, err := config.Build(); err != nil {
		panic(err)
	} else {
		zap.ReplaceGlobals(value)
	}

	if value, err := zapcore.ParseLevel(env.Get("LOG_LEVEL", level)); err != nil {
		panic(err)
	} else {
		atomicLevel.SetLevel(value)
	}
}

func NewProductionConfig() zap.Config {
	config = zap.NewProductionConfig()
	config.Encoding = env.Get("LOG_ENCODING", "json")
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return config
}

func NewDevelopmentConfig() zap.Config {
	config = zap.NewDevelopmentConfig()
	config.Encoding = env.Get("LOG_ENCODING", "console")
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
	return config
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
