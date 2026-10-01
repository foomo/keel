package keeltemporal

// see https://docs.temporal.io/go/error-handling/

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
)

// ActivityErrorType is the application error type used by [NewActivityError].
const ActivityErrorType = "keeltemporal.ActivityError"

// NewActivityError returns a non-retryable [temporal.ApplicationError] of
// type [ActivityErrorType] wrapping err with msg and optional details.
func NewActivityError(msg string, err error, details ...any) error {
	return temporal.NewNonRetryableApplicationError(msg, ActivityErrorType, err, details...)
}

// IsErrorType reports whether err wraps a [temporal.ApplicationError] of the
// given errorType.
func IsErrorType(err error, errorType string) bool {
	if applicationErr := AsApplicationError(err); applicationErr != nil && applicationErr.Type() == errorType {
		return true
	}

	return false
}

// IsActivityError reports whether err wraps an error created by
// [NewActivityError].
func IsActivityError(err error) bool {
	return IsErrorType(err, ActivityErrorType)
}

// IsApplicationError reports whether err wraps a [temporal.ApplicationError].
// The handler argument is ignored.
func IsApplicationError(err error, handler func(applicationErr *temporal.ApplicationError)) bool {
	return AsApplicationError(err) != nil
}

// AsApplicationError returns the first [temporal.ApplicationError] in err's
// chain, or nil if there is none.
func AsApplicationError(err error) *temporal.ApplicationError {
	var applicationErr *temporal.ApplicationError
	if err != nil && errors.As(err, &applicationErr) {
		return applicationErr
	}

	return nil
}

// IsCanceledError reports whether err wraps a [temporal.CanceledError].
// The handler argument is ignored.
func IsCanceledError(err error, handler func(canceledErr *temporal.CanceledError)) bool {
	return AsCanceledError(err) != nil
}

// AsCanceledError returns the first [temporal.CanceledError] in err's chain,
// or nil if there is none.
func AsCanceledError(err error) *temporal.CanceledError {
	var canceledErr *temporal.CanceledError
	if err != nil && errors.As(err, &canceledErr) {
		return canceledErr
	}

	return nil
}

// IsTimeoutError reports whether err wraps a [temporal.TimeoutError].
// The handler argument is ignored.
func IsTimeoutError(err error, handler func(timeoutErr *temporal.TimeoutError)) bool {
	return AsTimeoutError(err) != nil
}

// AsTimeoutError returns the first [temporal.TimeoutError] in err's chain,
// or nil if there is none.
func AsTimeoutError(err error) *temporal.TimeoutError {
	var timeoutErr *temporal.TimeoutError
	if err != nil && errors.As(err, &timeoutErr) {
		return timeoutErr
	}

	return nil
}

// IsPanicError reports whether err wraps a [temporal.PanicError].
// The handler argument is ignored.
func IsPanicError(err error, handler func(panicErr *temporal.PanicError)) bool {
	return AsPanicError(err) != nil
}

// AsPanicError returns the first [temporal.PanicError] in err's chain, or nil
// if there is none.
func AsPanicError(err error) *temporal.PanicError {
	var panicErr *temporal.PanicError
	if err != nil && errors.As(err, &panicErr) {
		return panicErr
	}

	return nil
}
