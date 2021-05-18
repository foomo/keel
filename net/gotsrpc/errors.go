package gotsrpc

// Error type
type Error string

// Common errors
const (
	ErrorNotFound         Error = "notFound"
	ErrorForbidden        Error = "forbidden"
	ErrorPermissionDenied Error = "permissionDenied"
)

// NewError returns a pointer error
func NewError(e Error) *Error {
	return &e
}

// Error interface
func (e *Error) Error() string {
	return string(*e)
}
