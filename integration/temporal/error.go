package keeltemporal

// ActivityCanceledError indicated that the activity is canceled on purpose
type ActivityCanceledError struct {
	error
}

func NewActivityCanceledError(err error) *ActivityCanceledError {
	return &ActivityCanceledError{
		error: err,
	}
}

// ActivityFailedError indicated that sth in the activity failed
type ActivityFailedError struct {
	error
}

func NewActivityFailedError(err error) *ActivityFailedError {
	return &ActivityFailedError{
		error: err,
	}
}

// ActivityUnprocessableError indicates that the activity is not
type ActivityUnprocessableError struct {
	error
}

func NewActivityUnprocessableError(err error) *ActivityUnprocessableError {
	return &ActivityUnprocessableError{
		error: err,
	}
}
