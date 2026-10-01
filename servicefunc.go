package keel

import (
	"context"
)

// ServiceFunc adapts a function to the [Service] interface.
type ServiceFunc func(ctx context.Context) error

// Start calls fn(ctx).
func (fn ServiceFunc) Start(ctx context.Context) error {
	return fn(ctx)
}
