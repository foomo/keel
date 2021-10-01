package store

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	EntityWithVersionsIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: "id", Value: 1},
			{Key: "version", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
)

// EntityWithVersions type
type EntityWithVersions struct {
	Version uint32 `json:"version" bson:"version"  yaml:"version"`
}

// GetVersion api implementation
func (e *EntityWithVersions) GetVersion() uint32 {
	return e.Version
}

// SetVersion api implementation
func (e *EntityWithVersions) SetVersion(value uint32) {
	e.Version = value
}

// IncreaseVersion api implementation
func (e *EntityWithVersions) IncreaseVersion() uint32 {
	e.Version = e.Version + 1
	return e.GetVersion()
}
