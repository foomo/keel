package keeltest

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func Logger(t zaptest.TestingT) *zap.Logger {
	var l *zap.Logger
	if t == nil {
		c := zap.NewDevelopmentConfig()
		c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		c.EncoderConfig.TimeKey = ""
		l, _ = c.Build()
	} else {
		l = zaptest.NewLogger(t)
	}
	// setup logger
	zap.ReplaceGlobals(l)
	return l
}
