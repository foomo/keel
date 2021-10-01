package keelmongo

import "time"

type Entity interface {
	SetID(string)
	GetID() string
}

type EntityWithVersion interface {
	GetVersion() uint32
	IncreaseVersion() uint32
}

type EntityWithTimestamps interface {
	SetCreatedAt(time.Time)
	GetCreatedAt() time.Time
	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time
}
