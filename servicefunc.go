package keel

import (
	"context"
)

type ServiceFunc func(ctx context.Context) error

func (fn ServiceFunc) Start(ctx context.Context) error {
	return fn(ctx)
}
