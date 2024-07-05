package keeltest

import (
	"context"
)

// Service interface
type Service interface {
	URL() string
	Name() string
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}
