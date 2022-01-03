package log

import (
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
	switch env.Get("LOG", ModeProd) {
	case ModeDev:
		atomicLevel.SetLevel(zap.DebugLevel)
		config = zap.Config{
			Level:            atomicLevel,
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
		config = zap.Config{
			Level:       atomicLevel,
			Development: false,
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	config.EncoderConfig.TimeKey = "time"
	config.DisableCaller = env.GetBool("LOG_DISABLE_STACKTRACE", true)
	config.DisableStacktrace = env.GetBool("LOG_DISABLE_CALLER", true)
	l, err := config.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
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
func Must(l *zap.Logger, err error, msg string) {
	if err != nil {
		if l == nil {
			l = Logger()
		}
		l.WithOptions(zap.AddCallerSkip(1)).Fatal(msg, FError(err))
	}
}
