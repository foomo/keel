package keel

import "context"

type ProbeType string

const probesServiceName = "probes"

const (
	Liveliness ProbeType = "liveliness"
	Readiness  ProbeType = "readiness"
	Startup    ProbeType = "startup"
)

type ProbeHandlers struct {
	handler   interface{}
	probeType string
}

// Health interface
type Health interface {
	Ping() bool
}

// HealthFn interface
type HealthFn func() bool

// ErrorHealthFn interface
type ErrorHealthFn func() (bool, error)

// ErrorHealth interface
type ErrorHealth interface {
	Ping() (bool, error)
}

// HealthWithContext interface
type HealthWithContext interface {
	Ping(ctx context.Context) bool
}

// HealthWithContextFn interface
type HealthWithContextFn func(ctx context.Context) bool

// ErrorHealthWithContext interface
type ErrorHealthWithContext interface {
	Ping(ctx context.Context) (bool, error)
}
