package main

import (
	"context"
	"time"

	keelpersistence "github.com/foomo/keel/persistence"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/foomo/keel"
	"github.com/foomo/keel/examples/persistence/mongo/repository"
	"github.com/foomo/keel/examples/persistence/mongo/store"
	"github.com/foomo/keel/log"
	keelmongo "github.com/foomo/keel/persistence/mongo"
)

// docker run -it --rm -p 27017:27017 mongo
func main() {
	svr := keel.NewServer()

	// get the logger
	l := svr.Logger()

	cDateTime := &store.DateTimeCodec{}
	rb := bson.NewRegistry()
	rb.RegisterTypeEncoder(store.TDateTime, cDateTime)
	rb.RegisterTypeDecoder(store.TDateTime, cDateTime)

	// create persistor
	persistor, err := keelmongo.New(
		svr.Context(),
		"mongodb://localhost:27017/dummy",
		// enable telemetry (enabled by default)
		keelmongo.WithOtelEnabled(true),
		keelmongo.WithClientOptions(
			options.Client().SetRegistry(rb),
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
	log.Must(l, repo.Insert(context.Background(), newEntity), "failed to insert")

	// fail insert for duplicate entity
	l.Info("Try to insert with duplicate key")
	if err := repo.Insert(context.Background(), &store.Dummy{
		Entity: store.NewEntity(newEntity.ID),
	}); mongo.IsDuplicateKeyError(err) {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else if err != nil {
		l.Error("unexpected error", log.FValue(err.Error()))
	} else {
		l.Error("unexpected success")
	}

	// fail insert for duplicate entity
	l.Info("Try to upsert with duplicate key")
	if err := repo.Upsert(context.Background(), &store.Dummy{
		Entity: store.NewEntity(newEntity.ID),
	}); mongo.IsDuplicateKeyError(err) {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else if err != nil {
		l.Error("unexpected error", log.FValue(err.Error()))
	} else {
		l.Error("unexpected success")
	}

	l.Info("Try to upsert many with duplicate key")
	if err := repo.UpsertMany(context.Background(), []*store.Dummy{{
		Entity: store.NewEntity(newEntity.ID),
	}}); mongo.IsDuplicateKeyError(err) {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else if err != nil {
		l.Error("unexpected error", log.FValue(err.Error()))
	} else {
		l.Error("unexpected success")
	}

	// get entity x2
	l.Info("Try to upsert with dirty write")
	newEntityA, err := repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")

	newEntityB, err := repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")

	// update entity A
	log.Must(l, repo.Upsert(context.Background(), newEntityA), "ERROR: failed to load new entity")

	// update entity B
	if err := repo.Upsert(context.Background(), newEntityB); errors.Is(err, keelpersistence.ErrDirtyWrite) {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else if err != nil {
		l.Error("unexpected error", log.FValue(err.Error()))
	} else {
		l.Error("unexpected success")
	}

	l.Info("Try to upsert many with dirty write")
	newEntityA, err = repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")

	newEntityB, err = repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")

	// update entity A
	log.Must(l, repo.UpsertMany(context.Background(), []*store.Dummy{newEntityA}), "ERROR: failed to load new entity")

	l.Info("Try to upsert many with dirty write")
	if err := repo.UpsertMany(context.Background(), []*store.Dummy{newEntityB}); errors.Is(err, keelpersistence.ErrDirtyWrite) {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else if err != nil {
		l.Error("unexpected error", log.FValue(err.Error()))
	} else {
		l.Error("unexpected success")
	}

	svr.Run()
}
