package store

import (
	"time"
)

// EntityWithTimestamps holds the creation and update times of a document.
type EntityWithTimestamps struct {
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"  yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"  yaml:"updatedAt"`
}

// GetCreatedAt returns the creation time.
func (e *EntityWithTimestamps) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// SetCreatedAt sets the creation time.
func (e *EntityWithTimestamps) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// GetUpdatedAt returns the update time.
func (e *EntityWithTimestamps) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// SetUpdatedAt sets the update time.
func (e *EntityWithTimestamps) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}
