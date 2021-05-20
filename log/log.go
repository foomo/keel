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

// logger holds the global logger
var logger *zap.Logger

func init() {
	var config zap.Config
	switch env.Get("LOG", ModeProd) {
	case ModeDev:
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
	default:
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	config.DisableStacktrace = env.GetBool("LOG_DISABLE_STACKTRACE", true)
	config.DisableCaller = env.GetBool("LOG_DISABLE_CALLER", true)
	if l, err := config.Build(); err != nil {
		panic(err)
	} else {
		logger = l
	}
}

// Logger return the logger instance
func Logger() *zap.Logger {
	return logger
}

func Sync() error {
	return logger.Sync()
}

func MustSync() {
	if err := logger.Sync(); err != nil {
		fmt.Println(err)
	}
}

// Must logs a fatal error if given
func Must(l *zap.Logger, err error, msg string) {
	if err != nil {
		if l == nil {
			l = Logger()
		}
		l.Fatal(msg, FError(err), FStackSkip(1))
	}
}
