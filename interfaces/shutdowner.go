package interfaces

import "context"

// Shutdowner interface
type Shutdowner interface {
	Shutdown()
}

// ErrorShutdowner interface
type ErrorShutdowner interface {
	Shutdown() error
}

// ShutdownerWithContext interface
type ShutdownerWithContext interface {
	Shutdown(ctx context.Context)
}

// ErrorShutdownerWithContext interface
type ErrorShutdownerWithContext interface {
	Shutdown(ctx context.Context) error
}
