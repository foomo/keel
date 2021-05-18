package keel

import (
	"context"
)

// Service interface
type Service interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}
