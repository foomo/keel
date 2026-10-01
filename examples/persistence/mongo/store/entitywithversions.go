package store

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EntityWithVersionsIndex is a unique index on the id and version fields.
var (
	EntityWithVersionsIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: "id", Value: 1},
			{Key: "version", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
)

// EntityWithVersions holds the version of a document used to detect dirty writes.
type EntityWithVersions struct {
	Version uint32 `json:"version" bson:"version"  yaml:"version"`
}

// GetVersion returns the version.
func (e *EntityWithVersions) GetVersion() uint32 {
	return e.Version
}

// SetVersion sets the version.
func (e *EntityWithVersions) SetVersion(value uint32) {
	e.Version = value
}

// IncreaseVersion increments the version and returns it.
func (e *EntityWithVersions) IncreaseVersion() uint32 {
	e.Version++
	return e.GetVersion()
}
