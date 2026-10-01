// Package keelmongo provides a MongoDB persistor for keel services built on
// the official go.mongodb.org/mongo-driver/v2 driver, with optional
// OpenTelemetry instrumentation.
//
// # Persistor and collections
//
// [New] connects to the database named in the connection URI. Collections
// are obtained via [Persistor.Collection], which can also ensure indexes:
//
//	p, err := keelmongo.New(ctx, "mongodb://localhost:27017/mydb")
//	col, err := p.Collection("users", keelmongo.CollectionWithIndexes(...))
//	err = col.Get(ctx, "id-1", &user)
//
// [Collection] stores documents keyed by an "id" field. Entities implementing
// [EntityWithTimestamps] get their timestamps maintained on write, and
// entities implementing [EntityWithVersion] get optimistic locking:
// conflicting writes fail with [keelpersistence.ErrDirtyWrite]. Lookups of
// missing documents fail with [keelpersistence.ErrNotFound].
//
// # Distributed lock
//
// [DistributedRWLock] implements [keelpersistence.DistributedRWLock] on a
// dedicated collection using github.com/foomo/mongo-lock.
//
// # Documentation
//
// [Readme] renders a Markdown table of all collections and named indexes
// created through this package.
package keelmongo
