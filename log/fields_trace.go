package log

import (
	"go.uber.org/zap"
)

const (
	TraceID = "trace_id"
)

func FTraceID(traceID string) zap.Field {
	return zap.String(TraceID, traceID)
}
