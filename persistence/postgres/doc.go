// Package keelpostgres provides a thin PostgreSQL persistor built on
// [database/sql] and the github.com/lib/pq driver.
//
// [New] opens and pings the database and optionally runs an initialization
// statement:
//
//	p, err := keelpostgres.New(ctx, "postgres://user:pass@localhost:5432/db",
//		keelpostgres.WithInit("create table if not exists ..."),
//	)
package keelpostgres
