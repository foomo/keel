package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/foomo/keel/example/persistence/mongo/store"
	keelmongo "github.com/foomo/keel/persistence/mongo"
)

type DummyRepository struct {
	collection *keelmongo.Collection
}

// NewDummyRepository constructor
func NewDummyRepository(collection *keelmongo.Collection) *DummyRepository {
	return &DummyRepository{
		collection: collection,
	}
}

// Get entity
func (r *DummyRepository) Get(ctx context.Context, id string, opts ...*options.FindOneOptions) (*store.Dummy, error) {
	var ret store.Dummy
	if err := r.collection.Get(ctx, id, &ret, opts...); err != nil {
		return nil, err
	}
	return &ret, nil
}

// Upsert entity
func (r *DummyRepository) Upsert(ctx context.Context, entity *store.Dummy) (err error) {
	if err := r.collection.Upsert(ctx, entity.GetID(), entity); err != nil {
		return err
	}
	return nil
}

// Delete entity
func (r *DummyRepository) Delete(ctx context.Context, id string) error {
	return r.collection.Delete(ctx, id)
}
