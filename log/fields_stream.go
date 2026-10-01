package log

import (
	"go.uber.org/zap"
)

const (
	// StreamQueueKey is the log field key "queue".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	StreamQueueKey = "queue"
	// StreamSubjectKey is the log field key "subject".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	StreamSubjectKey = "subject"
)

// FStreamQueue returns a field with the given value under [StreamQueueKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FStreamQueue(queue string) zap.Field {
	return zap.String(StreamQueueKey, queue)
}

// FStreamSubject returns a field with the given value under [StreamSubjectKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FStreamSubject(name string) zap.Field {
	return zap.String(StreamSubjectKey, name)
}
