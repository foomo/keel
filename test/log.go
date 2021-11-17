package keeltest

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func Logger(t zaptest.TestingT) *zap.Logger {
	l := zaptest.NewLogger(t)
	// setup logger
	zap.ReplaceGlobals(l)
	return l
}
