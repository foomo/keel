package keelpostgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	// Persistor wraps a [sql.DB] connected to PostgreSQL. It is exported so it
	// can be embedded into types in other packages. Its methods are safe for
	// concurrent use as far as the underlying [sql.DB] is.
	Persistor struct {
		db *sql.DB
		l  *zap.Logger
	}
	// Options configures [New].
	Options struct {
		// Init is an optional SQL statement executed once after connecting.
		Init string
		// Logger is the logger stored on the [Persistor].
		Logger *zap.Logger
	}
	// Option configures [Options].
	Option func(*Options)
)

// WithInit sets an SQL statement that [New] executes after connecting.
// Defaults to none.
func WithInit(v string) Option {
	return func(o *Options) {
		o.Init = v
	}
}

// WithLogger sets the logger. Defaults to [log.Logger].
func WithLogger(v *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = v
	}
}

// DefaultOptions returns the default [Options] using the global keel logger.
func DefaultOptions() Options {
	return Options{
		Logger: log.Logger(),
	}
}

// New connects to the PostgreSQL database identified by the connection
// string dns, verifies the connection with a ping and, if configured via
// [WithInit], executes the init statement. It returns an error if the
// connection string is invalid, the ping fails, or the init statement fails.
func New(ctx context.Context, dns string, opts ...Option) (*Persistor, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	connector, err := pq.NewConnector(dns)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)

	// TODO @franklin add telemetry

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	p := &Persistor{
		db: db,
		l:  o.Logger,
	}

	// initialize
	if o.Init != "" {
		if _, err := p.db.ExecContext(ctx, o.Init); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// TableExists reports whether a table with the given name exists in
// information_schema.tables.
func (p *Persistor) TableExists(ctx context.Context, name string) (bool, error) {
	var n int64
	if err := p.db.QueryRowContext(ctx, `select 1 from information_schema.tables where table_name=$1`, name).Scan(&n); errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// Ping verifies the database connection is alive.
func (p *Persistor) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// DB returns the underlying [sql.DB].
func (p *Persistor) DB() *sql.DB {
	return p.db
}

// Conn returns a single dedicated connection from the pool. The caller must
// close it when done.
func (p *Persistor) Conn(ctx context.Context) (*sql.Conn, error) {
	return p.db.Conn(ctx)
}

// Close closes the underlying database.
func (p *Persistor) Close() error {
	return p.db.Close()
}
