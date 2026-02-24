package store

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	EntityIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: "id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
)

// Entity type
type Entity struct {
	ID     string        `json:"id" bson:"id" yaml:"id"`
	BsonID bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" yaml:"_id,omitempty"` //nolint:tagliatelle
}

func NewEntity(id string) Entity {
	return Entity{ID: id}
}

// GetID api implementation
func (e *Entity) GetID() string {
	return e.ID
}

// SetID api implementation
func (e *Entity) SetID(value string) {
	e.ID = value
}
