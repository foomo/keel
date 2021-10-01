package keelerrors

import "errors"

type wrappedError struct {
	err   error
	cause error
}

func NewWrappedError(err, cause error) error {
	return &wrappedError{
		err:   err,
		cause: cause,
	}
}

func (e *wrappedError) Is(target error) bool {
	return errors.Is(e.err, target) || errors.Is(e.cause, target)
}

func (e *wrappedError) Cause() error {
	return e.cause
}

func (e *wrappedError) Unwrap() error {
	return e.err
}

func (e *wrappedError) Error() string {
	return e.err.Error() + ": " + e.cause.Error()
}
