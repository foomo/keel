package log

import (
	"go.uber.org/zap"
)

const (
	SpanID  = "span_id"
	TraceID = "trace_id"
)

func FSpanID(traceID string) zap.Field {
	return zap.String(SpanID, traceID)
}

func FTraceID(traceID string) zap.Field {
	return zap.String(TraceID, traceID)
}
