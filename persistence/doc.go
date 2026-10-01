// Package keelpersistence provides backend-agnostic persistence contracts
// shared by the keel storage integrations: the sentinel errors [ErrNotFound]
// and [ErrDirtyWrite], and the [DistributedRWLock] interface.
//
// Concrete implementations live in separate modules, such as
// github.com/foomo/keel/persistence/mongo and
// github.com/foomo/keel/persistence/postgres.
package keelpersistence
