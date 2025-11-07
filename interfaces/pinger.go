package interfaces

import "context"

// ErrorPinger interface
type ErrorPinger interface {
	Ping() error
}

// ErrorPingerWithContext interface
type ErrorPingerWithContext interface {
	Ping(ctx context.Context) error
}
