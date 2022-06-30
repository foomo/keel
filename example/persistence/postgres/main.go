package main

import (
	"github.com/davecgh/go-spew/spew"

	"github.com/foomo/keel"
	"github.com/foomo/keel/example/persistence/postgres/repository"
	"github.com/foomo/keel/log"
	keelpostgres "github.com/foomo/keel/persistence/postgres"
)

// docker run -it --rm -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres postgres:11-alpine
func main() {
	svr := keel.NewServer()

	// get the logger
	l := svr.Logger()

	// create persistor
	persistor, err := keelpostgres.New(
		svr.Context(),
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		keelpostgres.WithInit(`
			create table if not exists tasks (
				id serial primary key,
				description text not null
			);
	`),
	)
	// use log must helper to exit on error
	log.Must(l, err, "failed to create persistor")

	// ensure to add the persistor to the closers
	svr.AddClosers(persistor)

	repo := repository.NewTaskRepository(persistor)

	if value, err := repo.List(svr.Context()); err != nil {
		l.Error(err.Error())
	} else {
		spew.Dump(value)
	}

	if err := repo.Insert(svr.Context(), "one"); err != nil {
		l.Error(err.Error())
	}

	if value, err := repo.List(svr.Context()); err != nil {
		l.Error(err.Error())
	} else {
		spew.Dump(value)
	}

	svr.Run()
}
