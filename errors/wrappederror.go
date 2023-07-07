package keelerrors

type wrappedError struct {
	err   error
	cause error
}

// NewWrappedError returns a new wrapped error
func NewWrappedError(err, cause error) error {
	return &wrappedError{
		err:   err,
		cause: cause,
	}
}

func (e *wrappedError) Error() string {
	return e.err.Error() + ": " + e.cause.Error()
}

func (e *wrappedError) Unwrap() []error {
	return []error{e.err, e.cause}
}
