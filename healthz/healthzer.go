package healthz

import "context"

// healther adapts a function to the [ErrorHealthzWithContext] interface.
type healther struct {
	handle func(context.Context) error
}

// NewHealthzerFn returns a probe that calls handle on each health check.
// The returned value also implements Close(ctx) error by calling handle.
func NewHealthzerFn(handle func(context.Context) error) healther {
	return healther{
		handle: handle,
	}
}

// Healthz calls the wrapped function.
func (h healther) Healthz(ctx context.Context) error {
	return h.handle(ctx)
}

// Close calls the wrapped function.
func (h healther) Close(ctx context.Context) error {
	return h.handle(ctx)
}

// BoolHealthzer is a probe that reports whether it is healthy.
type BoolHealthzer interface {
	Healthz() bool
}

// BoolHealthzerWithContext is a probe that reports whether it is healthy,
// honoring the request context.
type BoolHealthzerWithContext interface {
	Healthz(ctx context.Context) bool
}

// ErrorHealthzer is a probe that returns a non-nil error when unhealthy.
type ErrorHealthzer interface {
	Healthz() error
}

// ErrorHealthzWithContext is a probe that returns a non-nil error when
// unhealthy, honoring the request context.
type ErrorHealthzWithContext interface {
	Healthz(ctx context.Context) error
}
