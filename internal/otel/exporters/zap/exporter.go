package zap

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Exporter implements [sdklog.Exporter].
var _ sdklog.Exporter = &Exporter{}

// Exporter writes OpenTelemetry log records to a zap logger, mapping the
// severity to a zap level and attributes, trace and span IDs to fields. It is
// safe for concurrent use. Exporter must be created with [New].
type Exporter struct {
	logger   *zap.Logger
	shutdown atomic.Bool
}

// New returns an [Exporter] writing to logger. The returned error is always
// nil.
func New(logger *zap.Logger) (*Exporter, error) {
	e := Exporter{
		logger: logger,
	}

	return &e, nil
}

// Export writes records to the logger. It stops and returns the context
// error if ctx is done, and does nothing after [Exporter.Shutdown].
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

// export writes a single record to the logger.
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

	r.WalkAttributes(func(kv attribute.KeyValue) bool {
		switch kv.Value.Type() {
		case attribute.BOOL:
			fields = append(fields, zap.Bool(string(kv.Key), kv.Value.AsBool()))
		case attribute.STRING:
			fields = append(fields, zap.String(string(kv.Key), kv.Value.AsString()))
		case attribute.FLOAT64:
			fields = append(fields, zap.Float64(string(kv.Key), kv.Value.AsFloat64()))
		case attribute.INT64:
			fields = append(fields, zap.Int64(string(kv.Key), kv.Value.AsInt64()))
		default:
			fields = append(fields, zap.Any(string(kv.Key), kv.Value))
		}

		return true
	})

	e.logger.Log(convertLevel(r.Severity()), r.Body().String(), fields...)
}

// convertLevel maps an OpenTelemetry severity to a zap level. Only the base
// severities (e.g. SeverityInfo, not SeverityInfo2) are mapped; all others
// return zapcore.InvalidLevel.
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
