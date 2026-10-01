package log

import (
	"go.uber.org/zap"
)

// Log field keys for trace correlation.
const (
	SpanID  = "span_id"
	TraceID = "trace_id"
)

// FSpanID creates a zap.Field with the given span id under the key "span_id".
func FSpanID(traceID string) zap.Field {
	return zap.String(SpanID, traceID)
}

// FTraceID creates a zap.Field with the given trace id under the key "trace_id".
func FTraceID(traceID string) zap.Field {
	return zap.String(TraceID, traceID)
}
