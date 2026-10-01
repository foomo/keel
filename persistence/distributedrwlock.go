package keelpersistence

import (
	"context"
)

// DistributedRWLock coordinates shared and exclusive holders of a single
// cluster-wide resource. Any number of shared holders may hold the lock
// concurrently, while an exclusive holder excludes all others.
//
// AcquireShared and AcquireExclusive return (true, nil) on success and
// (false, nil) when blocked by the other side; a non-nil error indicates a
// backend failure. Release frees the lock held under lockID.
//
// See github.com/foomo/keel/persistence/mongo for a MongoDB implementation.
type DistributedRWLock interface {
	AcquireShared(ctx context.Context, lockID string) (bool, error)
	AcquireExclusive(ctx context.Context, lockID string) (bool, error)
	Release(ctx context.Context, lockID string) error
}
