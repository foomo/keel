package healthz

import "context"

type healther struct {
	handle func(context.Context) error
}

func NewHealthzerFn(handle func(context.Context) error) healther {
	return healther{
		handle: handle,
	}
}

func (h healther) Healthz(ctx context.Context) error {
	return h.handle(ctx)
}

func (h healther) Close(ctx context.Context) error {
	return h.handle(ctx)
}

// BoolHealthzer interface
type BoolHealthzer interface {
	Healthz() bool
}

// BoolHealthzerWithContext interface
type BoolHealthzerWithContext interface {
	Healthz(ctx context.Context) bool
}

// ErrorHealthzer interface
type ErrorHealthzer interface {
	Healthz() error
}

// ErrorHealthzWithContext interface
type ErrorHealthzWithContext interface {
	Healthz(ctx context.Context) error
}
