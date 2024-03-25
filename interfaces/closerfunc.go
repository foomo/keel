package interfaces

import (
	"context"
)

type CloserFunc func(ctx context.Context) error

func (f CloserFunc) Close(ctx context.Context) error {
	return f(ctx)
}
