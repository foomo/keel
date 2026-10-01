package keeltest

import (
	"context"
)

// Service is a service managed by a [Server]. URL returns the address the
// service is reachable at once started.
type Service interface {
	URL() string
	Name() string
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}
