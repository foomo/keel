package keel

import "context"

type Probes map[ProbeType][]interface{}

// BoolProbeFn interface
type BoolProbeFn func() bool

// ErrorProbeFn interface
type ErrorProbeFn func() error

// BoolProbeWithContextFn interface
type BoolProbeWithContextFn func(context.Context) bool

// ErrorProbeWithContextFn interface
type ErrorProbeWithContextFn func(context.Context) error

// ErrorPingProbe interface
type ErrorPingProbe interface {
	Ping() error
}

// ErrorPingProbeWithContext interface
type ErrorPingProbeWithContext interface {
	Ping(context.Context) error
}
