package keelgotsrpc

// Error type
// Deprecated: use standard errors instead
type Error string

// Common errors
const (
	// Deprecated: use standard errors instead
	ErrorNotFound Error = "notFound"
	// Deprecated: use standard errors instead
	ErrorForbidden Error = "forbidden"
	// Deprecated: use standard errors instead
	ErrorPermissionDenied Error = "permissionDenied"
)

// NewError returns a pointer error
// Deprecated: use standard errors instead
func NewError(e Error) *Error {
	return &e
}

// Is interface
// Deprecated: use standard errors instead
func (e *Error) Is(err error) bool {
	if e == nil || err == nil {
		return false
	} else if v, ok := err.(*Error); ok && v != nil {
		return e.Error() == v.Error()
	}
	return false
}

// Error interface
// Deprecated: use standard errors instead
func (e *Error) Error() string {
	return string(*e)
}
