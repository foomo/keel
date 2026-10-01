package store

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EntityIndex is a unique index on the id field.
var (
	EntityIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: "id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
)

// Entity holds the ID and the MongoDB object ID of a document.
type Entity struct {
	ID     string        `json:"id" bson:"id" yaml:"id"`
	BsonID bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" yaml:"_id,omitempty"` //nolint:tagliatelle
}

// NewEntity returns an Entity with the given id.
func NewEntity(id string) Entity {
	return Entity{ID: id}
}

// GetID returns the ID.
func (e *Entity) GetID() string {
	return e.ID
}

// SetID sets the ID.
func (e *Entity) SetID(value string) {
	e.ID = value
}
