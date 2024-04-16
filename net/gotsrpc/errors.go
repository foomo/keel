package keelgotsrpc

// Error type
type Error string

// Common errors
const (
	ErrorNotFound         Error = "notFound"         //nolint:errname
	ErrorForbidden        Error = "forbidden"        //nolint:errname
	ErrorPermissionDenied Error = "permissionDenied" //nolint:errname
)

// NewError returns a pointer error
func NewError(e Error) *Error {
	return &e
}

// Is interface
func (e *Error) Is(err error) bool {
	if e == nil || err == nil {
		return false
	} else if v, ok := err.(*Error); ok && v != nil {
		return e.Error() == v.Error()
	}
	return false
}

// Error interface
func (e *Error) Error() string {
	return string(*e)
}
