package keelpostgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"

	"github.com/pkg/errors"
)

// Persistor exported to used also for embedding into other types in foreign packages.
type (
	Persistor struct {
		db *pgx.Conn
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

func New(ctx context.Context, uri string, opts ...Option) (*Persistor, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"

	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	config, err := pgx.ParseConfig(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uri")
	}

	config.Logger = zapadapter.NewLogger(o.Logger)
	// TODO @franklin performance config

	db, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect")
	}

	// TODO @franklin add telemetry

	if err := db.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	p := &Persistor{
		db: db,
		l:  o.Logger,
	}

	// initialize
	if o.Init != "" {
		if _, err := p.db.Exec(ctx, o.Init); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Persistor) TableExists(ctx context.Context, name string) (bool, error) {
	var n int64
	if err := p.db.QueryRow(ctx, `select 1 from information_schema.tables where table_name=$1`, name).Scan(&n); errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (p *Persistor) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}

func (p *Persistor) Conn() *pgx.Conn {
	return p.db
}

func (p *Persistor) Close(ctx context.Context) error {
	return p.db.Close(ctx)
}
