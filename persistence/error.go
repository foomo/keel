package keelpersistence

import (
	"github.com/pkg/errors"
)

var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("not found error")
	// ErrDirtyWrite is returned when an optimistic, version-checked write
	// fails because the stored entity was modified concurrently.
	ErrDirtyWrite = errors.New("dirty write error")
)
