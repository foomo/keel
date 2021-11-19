package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/foomo/keel"
	"github.com/foomo/keel/example/persistence/mongo/repository"
	"github.com/foomo/keel/example/persistence/mongo/store"
	"github.com/foomo/keel/log"
	keelmongo "github.com/foomo/keel/persistence/mongo"
)

// docker run -it --rm -p 27017:27017 mongo
func main() {
	svr := keel.NewServer()

	// get the logger
	l := svr.Logger()

	cDateTime := &store.DateTimeCodec{}
	rb := bson.NewRegistryBuilder()
	rb.RegisterCodec(store.TDateTime, cDateTime)

	// create persistor
	persistor, err := keelmongo.New(
		svr.Context(),
		"mongodb://localhost:27017/dummy",
		// enable telemetry (enabled by default)
		keelmongo.WithOtelEnabled(true),
		keelmongo.WithClientOptions(
			options.Client().SetRegistry(rb.Build()),
		),
	)
	// use log must helper to exit on error
	log.Must(l, err, "failed to create persistor")

	// ensure to add the persistor to the closers
	svr.AddClosers(persistor)

	// create repositories
	col, err := persistor.Collection(
		"dummy",
		// define indexes but beware of changes on large dbs
		keelmongo.CollectionWithIndexes(
			store.EntityIndex,
			store.EntityWithVersionsIndex,
		),
		// define max time for index creation
		keelmongo.CollectionWithIndexesMaxTime(time.Second*10),
	)
	log.Must(l, err, "failed to create collection")
	repo := repository.NewDummyRepository(col)

	// --- version example ---

	// insert entity
	newEntity := &store.Dummy{
		Entity: store.NewEntity(uuid.New().String()),
	}
	log.Must(l, repo.Upsert(context.Background(), newEntity), "failed to insert")

	// fail insert for duplicate entity
	duplicateEntity := &store.Dummy{
		Entity: store.NewEntity(newEntity.ID),
	}
	if err := repo.Upsert(context.Background(), duplicateEntity); err != nil {
		l.Info("OK: expected error", log.FValue(err.Error()))
	}

	// get entity x2
	newEntityA, err := repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")

	newEntityB, err := repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")

	// update entity A
	if err := repo.Upsert(context.Background(), newEntityA); err != nil {
		l.Error("ERROR: failed to load new entity")
	}
	// update entity B
	if err := repo.Upsert(context.Background(), newEntityB); err != nil {
		l.Info("OK: expected error", log.FValue(err.Error()))
	}

	svr.Run()
}
