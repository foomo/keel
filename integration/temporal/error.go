package keeltemporal

// see https://docs.temporal.io/go/error-handling/

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
)

const ActivityErrorType = "keeltemporal.ActivityError"

func NewActivityError(msg string, err error, details ...interface{}) error {
	return temporal.NewNonRetryableApplicationError(msg, ActivityErrorType, err, details...)
}

func IsErrorType(err error, errorType string) bool {
	if applicationErr := AsApplicationError(err); applicationErr != nil && applicationErr.Type() == errorType {
		return true
	}
	return false
}

func IsActivityError(err error) bool {
	return IsErrorType(err, ActivityErrorType)
}

func IsApplicationError(err error, handler func(applicationErr *temporal.ApplicationError)) bool {
	return AsApplicationError(err) != nil
}

func AsApplicationError(err error) *temporal.ApplicationError {
	var applicationErr *temporal.ApplicationError
	if err != nil && errors.As(err, &applicationErr) {
		return applicationErr
	}
	return nil
}

func IsCanceledError(err error, handler func(canceledErr *temporal.CanceledError)) bool {
	return AsCanceledError(err) != nil
}

func AsCanceledError(err error) *temporal.CanceledError {
	var canceledErr *temporal.CanceledError
	if err != nil && errors.As(err, &canceledErr) {
		return canceledErr
	}
	return nil
}

func IsTimeoutError(err error, handler func(timeoutErr *temporal.TimeoutError)) bool {
	return AsTimeoutError(err) != nil
}

func AsTimeoutError(err error) *temporal.TimeoutError {
	var timeoutErr *temporal.TimeoutError
	if err != nil && errors.As(err, &timeoutErr) {
		return timeoutErr
	}
	return nil
}

func IsPanicError(err error, handler func(panicErr *temporal.PanicError)) bool {
	return AsPanicError(err) != nil
}

func AsPanicError(err error) *temporal.PanicError {
	var panicErr *temporal.PanicError
	if err != nil && errors.As(err, &panicErr) {
		return panicErr
	}
	return nil
}
