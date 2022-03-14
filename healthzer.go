package keel

import "context"

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
