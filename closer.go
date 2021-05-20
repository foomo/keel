package keel

import "context"

// Syncer interface
type Syncer interface {
	Sync() error
}

// SyncerWithContext interface
type SyncerWithContext interface {
	Sync(ctx context.Context) error
}

// Closer interface
type Closer interface {
	Close() error
}

// CloserWithContext interface
type CloserWithContext interface {
	Close(ctx context.Context) error
}

// Shutdowner interface
type Shutdowner interface {
	Shutdown() error
}

// ShutdownerWithContext interface
type ShutdownerWithContext interface {
	Shutdown(ctx context.Context) error
}
