package log

import (
	"go.uber.org/zap"
)

const (
	// Deprecated: use semconv messaging attributes instead.
	StreamQueueKey = "queue"
	// Deprecated: use semconv messaging attributes instead.
	StreamSubjectKey = "subject"
)

// Deprecated: use semconv messaging attributes instead.
func FStreamQueue(queue string) zap.Field {
	return zap.String(StreamQueueKey, queue)
}

// Deprecated: use semconv messaging attributes instead.
func FStreamSubject(name string) zap.Field {
	return zap.String(StreamSubjectKey, name)
}
