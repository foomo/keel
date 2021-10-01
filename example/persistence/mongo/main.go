package main

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/foomo/keel"
	"github.com/foomo/keel/example/persistence/mongo/repository"
	"github.com/foomo/keel/example/persistence/mongo/store"
	"github.com/foomo/keel/log"
	keelmongo "github.com/foomo/keel/persistence/mongo"

	"github.com/google/uuid"
)

func main() {
	svr := keel.NewServer()

	// get the logger
	l := svr.Logger()

	// create persistor
	persistor, err := keelmongo.New(
		svr.Context(),
		"mongodb://localhost:27017/dummy",
		// enable telemetry (enabled by default)
		keelmongo.WithOtelEnabled(true),
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



	spew.Dump("=== 0 =======================")

	newEntity := &store.Dummy{
		Entity: store.NewEntity(uuid.New().String()),
	}
	log.Must(l, repo.Upsert(context.Background(), newEntity), "failed to insert")

	duplicateEntity := &store.Dummy{
		Entity: store.NewEntity(newEntity.ID),
	}
	if err := repo.Upsert(context.Background(), duplicateEntity); err != nil {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else {
		l.Error("missing duplicate error")
	}

	newEntityA, err := repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")
	spew.Dump(newEntityA.ID, newEntityA.Version)

	newEntityB, err := repo.Get(context.Background(), newEntity.ID)
	log.Must(l, err, "failed to load new entity")
	spew.Dump(newEntityB.ID, newEntityB.Version)

	spew.Dump("=== 1 =======================")

	if err := repo.Upsert(context.Background(), newEntityA); err != nil {
		l.Error("ERROR: failed to load new entity")
	} else {
		l.Info("updated newEntityA")
		spew.Dump(newEntityA.ID, newEntityA.Version)
		spew.Dump(newEntityB.ID, newEntityB.Version)
	}
	if err := repo.Upsert(context.Background(), newEntityB); err != nil {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else {
		l.Error("ERROR: missing dirty write error")
		spew.Dump(newEntityA.ID, newEntityA.Version)
		spew.Dump(newEntityB.ID, newEntityB.Version)
	}

	spew.Dump("=== 2 =======================")

	if err := repo.Upsert(context.Background(), newEntityA); err != nil {
		l.Error("ERROR: failed to load new entity")
	} else {
		l.Info("updated newEntityA")
	}
	if err := repo.Upsert(context.Background(), newEntityB); err != nil {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else {
		l.Error("ERROR: missing dirty write error")
		spew.Dump(newEntityA.ID, newEntityA.Version)
		spew.Dump(newEntityB.ID, newEntityB.Version)
	}

	spew.Dump("=== 3 =======================")

	if err := repo.Upsert(context.Background(), newEntityA); err != nil {
		l.Error("ERROR: failed to load new entity")
	} else {
		l.Info("updated newEntityA")
	}
	if err := repo.Upsert(context.Background(), newEntityB); err != nil {
		l.Info("OK: expected error", log.FValue(err.Error()))
	} else {
		l.Error("ERROR: missing dirty write error")
		spew.Dump(newEntityA.ID, newEntityA.Version)
		spew.Dump(newEntityB.ID, newEntityB.Version)
	}

	svr.Run()
}
