package log

import (
	"go.uber.org/zap"
)

const (
	StreamQueueKey   = "queue"
	StreamSubjectKey = "subject"
)

func FStreamQueue(queue string) zap.Field {
	return zap.String(StreamQueueKey, queue)
}

func FStreamSubject(name string) zap.Field {
	return zap.String(StreamSubjectKey, name)
}
