package keelmongo

import "time"

// Entity is a document identified by a string ID stored in the "id" field.
type Entity interface {
	SetID(id string)
	GetID() string
}

// EntityWithVersion is implemented by entities that use optimistic locking.
// The version is stored in the "version" field and IncreaseVersion
// increments it before each write.
type EntityWithVersion interface {
	GetVersion() uint32
	IncreaseVersion() uint32
}

// EntityWithTimestamps is implemented by entities whose creation and update
// times are maintained on write. The creation time is only set if zero.
type EntityWithTimestamps interface {
	SetCreatedAt(t time.Time)
	GetCreatedAt() time.Time
	SetUpdatedAt(t time.Time)
	GetUpdatedAt() time.Time
}
