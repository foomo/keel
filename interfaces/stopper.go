package interfaces

import "context"

// Stopper interface
type Stopper interface {
	Stop()
}

// ErrorStopper interface
type ErrorStopper interface {
	Stop() error
}

// StopperWithContext interface
type StopperWithContext interface {
	Stop(ctx context.Context)
}

// ErrorStopperWithContext interface
type ErrorStopperWithContext interface {
	Stop(ctx context.Context) error
}
