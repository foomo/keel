package log

import (
	"go.uber.org/zap"
)

const (
	ErrorTypeKey       = "error_type"
	ErrorMessageKey    = "error_message"
	ErrorStacktraceKey = "error_stacktrace"
)

func FError(err error) zap.Field {
	return zap.NamedError(ErrorMessageKey, err)
}

func FErrorType(errType string) zap.Field {
	return zap.String(ErrorTypeKey, errType)
}

func FStackSkip(skip int) zap.Field {
	return zap.StackSkip(ErrorStacktraceKey, skip)
}
