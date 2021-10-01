package store

import (
	"time"
)

type EntityWithTimestamps struct {
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"  yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"  yaml:"updatedAt"`
}

// GetCreatedAt api implementation
func (e *EntityWithTimestamps) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// SetCreatedAt api implementation
func (e *EntityWithTimestamps) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// GetUpdatedAt api implementation
func (e *EntityWithTimestamps) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// SetUpdatedAt api implementation
func (e *EntityWithTimestamps) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}
