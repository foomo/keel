package keel

import (
	"context"
)

// Service interface
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}
