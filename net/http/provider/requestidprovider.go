package provider

import (
	"github.com/google/uuid"
)

type RequestID func() string

// DefaultRequestID function
func DefaultRequestID() string {
	return uuid.New().String()
}
