package zap

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ sdklog.Exporter = &Exporter{}

// Exporter writes JSON-encoded log records to an [io.Writer] ([os.Stdout] by default).
// Exporter must be created with [New].
type Exporter struct {
	logger   *zap.Logger
	shutdown atomic.Bool
}

// New creates an [Exporter].
func New(logger *zap.Logger) (*Exporter, error) {
	e := Exporter{
		logger: logger,
	}

	return &e, nil
}

// Export exports log records to writer.
func (e *Exporter) Export(ctx context.Context, records []sdklog.Record) error {
	if e.shutdown.Load() {
		return nil
	}

	for _, record := range records {
		// Honor context cancellation.
		if err := ctx.Err(); err != nil {
			return err
		}

		e.export(record)
	}

	return nil
}

// Shutdown shuts down the Exporter.
// Calls to Export will perform no operation after this is called.
func (e *Exporter) Shutdown(context.Context) error {
	e.shutdown.Store(true)
	return nil
}

// ForceFlush performs no action.
func (e *Exporter) ForceFlush(context.Context) error {
	return nil
}

func (e *Exporter) export(r sdklog.Record) {
	var fields []zap.Field

	if v := r.EventName(); v != "" {
		fields = append(fields, zap.String("eventName", v))
	}

	if r.TraceID().IsValid() {
		fields = append(fields, zap.String("traceId", r.TraceID().String()))
	}

	if r.SpanID().IsValid() {
		fields = append(fields, zap.String("spanId", r.SpanID().String()))
	}

	r.WalkAttributes(func(kv log.KeyValue) bool {
		switch kv.Value.Kind() {
		case log.KindBool:
			fields = append(fields, zap.Bool(kv.Key, kv.Value.AsBool()))
		case log.KindString:
			fields = append(fields, zap.String(kv.Key, kv.Value.AsString()))
		case log.KindFloat64:
			fields = append(fields, zap.Float64(kv.Key, kv.Value.AsFloat64()))
		case log.KindInt64:
			fields = append(fields, zap.Int64(kv.Key, kv.Value.AsInt64()))
		default:
			fields = append(fields, zap.Any(kv.Key, kv.Value))
		}

		return true
	})

	e.logger.Log(convertLevel(r.Severity()), r.Body().String(), fields...)
}

func convertLevel(level log.Severity) zapcore.Level {
	switch level {
	case log.SeverityDebug:
		return zapcore.DebugLevel
	case log.SeverityInfo:
		return zapcore.InfoLevel
	case log.SeverityWarn:
		return zapcore.WarnLevel
	case log.SeverityError:
		return zapcore.ErrorLevel
	case log.SeverityFatal1:
		return zapcore.DPanicLevel
	case log.SeverityFatal2:
		return zapcore.PanicLevel
	case log.SeverityFatal3:
		return zapcore.FatalLevel
	default:
		return zapcore.InvalidLevel
	}
}
