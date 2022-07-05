package keelpostgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// Persistor exported to used also for embedding into other types in foreign packages.
type (
	Persistor struct {
		db *sql.DB
		l  *zap.Logger
	}
	Options struct {
		Init   string
		Logger *zap.Logger
	}
	Option func(*Options)
)

func WithInit(v string) Option {
	return func(o *Options) {
		o.Init = v
	}
}

func WithLogger(v *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = v
	}
}

func DefaultOptions() Options {
	return Options{
		Logger: log.Logger(),
	}
}

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

func (p *Persistor) TableExists(ctx context.Context, name string) (bool, error) {
	var n int64
	if err := p.db.QueryRowContext(ctx, `select 1 from information_schema.tables where table_name=$1`, name).Scan(&n); errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (p *Persistor) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Persistor) DB() *sql.DB {
	return p.db
}

func (p *Persistor) Conn(ctx context.Context) (*sql.Conn, error) {
	return p.db.Conn(ctx)
}

func (p *Persistor) Close() error {
	return p.db.Close()
}
