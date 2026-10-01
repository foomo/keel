package log

import (
	"reflect"

	"go.uber.org/zap"
)

// Log field keys for errors.
const (
	ErrorTypeKey       = "error_type"
	ErrorMessageKey    = "error_message"
	ErrorStacktraceKey = "error_stacktrace"
)

// FError creates a zap.Field with err under the key "error_message". A nil err
// yields a no-op field.
func FError(err error) zap.Field {
	return zap.NamedError(ErrorMessageKey, err)
}

// FErrorType creates a zap.Field with the dynamic type name of err under the
// key "error_type". It panics if err is nil.
func FErrorType(err error) zap.Field {
	return zap.String(ErrorTypeKey, reflect.TypeOf(err).String())
}

// FStackSkip creates a zap.Field with the current stack trace under the key
// "error_stacktrace", skipping the given number of frames.
func FStackSkip(skip int) zap.Field {
	return zap.StackSkip(ErrorStacktraceKey, skip)
}
