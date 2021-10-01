package keelpersistence

import (
	"github.com/pkg/errors"
)

var (
	ErrNotFound   = errors.New("not found error")
	ErrDirtyWrite = errors.New("dirty write error")
)
