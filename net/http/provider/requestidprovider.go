package provider

import (
	"github.com/google/uuid"
)

// RequestID returns a new request ID.
type RequestID func() string

// DefaultRequestID returns a random (version 4) UUID string.
func DefaultRequestID() string {
	return uuid.New().String()
}
