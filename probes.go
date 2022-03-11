package keel

import "context"

type Probes map[ProbeType][]interface{}

// Healthz

// BoolHealthz interface
type BoolHealthz interface {
	Healthz() bool
}

// BoolHealthzWithContext interface
type BoolHealthzWithContext interface {
	Healthz(ctx context.Context) bool
}

// ErrorHealthz interface
type ErrorHealthz interface {
	Healthz() error
}

// ErrorHealthzWithContext interface
type ErrorHealthzWithContext interface {
	Healthz(ctx context.Context) error
}

// Ping

// ErrorPingProbe interface
type ErrorPingProbe interface {
	Ping() error
}

// ErrorPingProbeWithContext interface
type ErrorPingProbeWithContext interface {
	Ping(context.Context) error
}
