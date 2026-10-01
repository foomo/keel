package keelmongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	mongolock "github.com/foomo/mongo-lock"
)

// DistributedRWLock implements [keelpersistence.DistributedRWLock] on a
// MongoDB collection, acquiring shared and exclusive locks on one configured
// resource.
type DistributedRWLock struct {
	client       *mongolock.Client
	resourceName string
	sharedTTL    uint
	exclusiveTTL uint
	drainWait    time.Duration
	drainTimeout time.Duration
}

// NewDistributedRWLock builds the lock on the configured collection of
// persistor. It returns an error if cfg lacks CollectionName or ResourceName.
// The lock's indexes are not created; call
// [DistributedRWLock.CreateIndexes] to ensure they exist.
func NewDistributedRWLock(persistor *Persistor, cfg DistributedRWLockConfig) (*DistributedRWLock, error) {
	cfg, err := cfg.normalized()
	if err != nil {
		return nil, err
	}

	col, err := persistor.Collection(cfg.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("distributedreadwritelock: collection: %w", err)
	}

	return &DistributedRWLock{
		client:       mongolock.NewClient(col.Col()),
		resourceName: cfg.ResourceName,
		sharedTTL:    uint(cfg.SharedTTL.Seconds()),
		exclusiveTTL: uint(cfg.ExclusiveTTL.Seconds()),
		drainWait:    cfg.DrainWait,
		drainTimeout: cfg.DrainTimeout,
	}, nil
}

// CreateIndexes creates the indexes required by the lock collection.
func (l *DistributedRWLock) CreateIndexes(ctx context.Context) error {
	return l.client.CreateIndexes(ctx)
}

// AcquireShared takes a shared lock. It returns (false, nil) when the exclusive
// lock is held — the caller should reject the request — and (true, nil) on
// success. Multiple shared holders are allowed concurrently.
func (l *DistributedRWLock) AcquireShared(ctx context.Context, lockID string) (bool, error) {
	err := l.client.SLock(ctx, l.resourceName, lockID, mongolock.LockDetails{TTL: l.sharedTTL}, -1)

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, mongolock.ErrAlreadyLocked):
		return false, nil
	default:
		return false, fmt.Errorf("distributedreadwritelock: shared: %w", err)
	}
}

// AcquireExclusive takes the exclusive lock. It retries every DrainWait for up
// to DrainTimeout to let in-flight shared locks drain, then returns
// (false, nil) if still blocked so the caller skips this round. It returns
// ctx.Err() if ctx is done while waiting.
func (l *DistributedRWLock) AcquireExclusive(ctx context.Context, lockID string) (bool, error) {
	deadline := time.Now().Add(l.drainTimeout)

	for {
		err := l.client.XLock(ctx, l.resourceName, lockID, mongolock.LockDetails{TTL: l.exclusiveTTL})
		if err == nil {
			return true, nil
		}

		if !errors.Is(err, mongolock.ErrAlreadyLocked) {
			return false, fmt.Errorf("distributedreadwritelock: exclusive: %w", err)
		}

		if time.Now().After(deadline) {
			return false, nil
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(l.drainWait):
		}
	}
}

// Release releases the lock held under lockID (shared or exclusive).
func (l *DistributedRWLock) Release(ctx context.Context, lockID string) error {
	if _, err := l.client.Unlock(ctx, lockID); err != nil {
		return fmt.Errorf("distributedreadwritelock: unlock: %w", err)
	}

	return nil
}
