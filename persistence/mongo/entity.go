package keelmongo

import "time"

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
