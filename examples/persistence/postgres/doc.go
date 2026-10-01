// Postgres is an example service demonstrating the PostgreSQL persistor with an
// init statement and a simple repository.
//
// Start PostgreSQL and run the example:
//
//	docker run -it --rm -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres postgres:11-alpine
//	go run ./examples/persistence/postgres
package main
