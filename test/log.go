package keeltest

import (
	"bytes"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

type (
	Logger struct {
		zap    *zap.Logger
		writer *testingWriter
	}
	LoggerOptions struct {
		Level      zapcore.LevelEnabler
		zapOptions []zap.Option
	}
	LoggerOption func(*LoggerOptions)
)

// LoggerWithLevel adds zap.Option's to a test Logger built by NewLogger.
func LoggerWithLevel(o zapcore.LevelEnabler) LoggerOption {
	return func(v *LoggerOptions) {
		v.Level = o
	}
}

// LoggerWithZapOptions adds zap.Option's to a test Logger built by NewLogger.
func LoggerWithZapOptions(o ...zap.Option) LoggerOption {
	return func(v *LoggerOptions) {
		v.zapOptions = o
	}
}

func NewLogger(t zaptest.TestingT, opts ...LoggerOption) *Logger {
	cfg := LoggerOptions{
		Level: zapcore.DebugLevel,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	writer := newTestingWriter(t)
	zapOptions := []zap.Option{
		zap.AddCaller(),
		// zap.AddCallerSkip(1),
		// Send zap errors to the same writer and mark the test as failed if that happens.
		zap.ErrorOutput(writer),
	}
	zapOptions = append(zapOptions, cfg.zapOptions...)

	e := zap.NewDevelopmentEncoderConfig()
	e.TimeKey = ""
	e.EncodeLevel = zapcore.CapitalColorLevelEncoder
	e.EncodeCaller = func(s zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(37), s.TrimmedPath()))
	}
	return &Logger{
		zap: zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(e),
				writer,
				cfg.Level,
			),
			zapOptions...,
		),
		writer: writer,
	}
}

func (l *Logger) T(t zaptest.TestingT) *Logger {
	l.writer.t = t
	return l
}

func (l *Logger) Zap() *zap.Logger {
	return l.zap
}

// testingWriter is a WriteSyncer that writes to the given testing.TB.
type testingWriter struct {
	t zaptest.TestingT
}

func newTestingWriter(t zaptest.TestingT) *testingWriter {
	return &testingWriter{t: t}
}

func (w *testingWriter) Write(p []byte) (n int, err error) {
	if w.t == nil {
		return fmt.Printf("%s", p)
	} else {
		// Note: t.Log is safe for concurrent use.
		// Strip trailing newline because t.Log always adds one.
		w.t.Logf("%s", bytes.TrimRight(p, "\n"))
		return len(p), nil
	}
}

func (w *testingWriter) Sync() error {
	return nil
}
