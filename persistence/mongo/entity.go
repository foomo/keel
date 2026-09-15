package keelmongo

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Entity interface {
	SetID(id string)
	GetID() string
}

type EntityWithVersion interface {
	GetVersion() uint32
	IncreaseVersion() uint32
}

type EntityWithTimestamps interface {
	SetCreatedAt(t time.Time)
	GetCreatedAt() time.Time
	SetUpdatedAt(t time.Time)
	GetUpdatedAt() time.Time
}

// EntityWithInsertOnlyFields is implemented by entities carrying fields that must
// only ever be written when the document is first inserted - typically a creation
// timestamp. Upsert and UpsertMany move the returned elements into $setOnInsert,
// so an upsert that turns out to be an update leaves them untouched.
//
// This keeps write-once semantics correct for blind writers: a caller that never
// read the document cannot supply the stored creation timestamp, and without this
// the entity's zero value would overwrite it on every re-import.
//
// Implementations must return no elements whenever the entity already carries the
// values itself. Those values then travel in $set as usual, and returning them
// here as well would make mongo reject the write with a conflicting update path.
// Fields returned here must be tagged `bson:",omitempty"` so that their zero value
// stays out of $set and cannot collide with $setOnInsert.
//
// Entities that do not implement this interface are written exactly as before.
type EntityWithInsertOnlyFields interface {
	InsertOnlyFields(now time.Time) bson.D
}
