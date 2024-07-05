package keelmongo

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"

	"github.com/foomo/keel/env"
)

// Persistor exported to used also for embedding into other types in foreign packages.
type (
	Persistor struct {
		client *mongo.Client
		db     *mongo.Database
	}
	Options struct {
		OtelEnabled     bool
		OtelOptions     []otelmongo.Option
		ClientOptions   []ClientOption
		DatabaseOptions []DatabaseOption
	}
	Option         func(o *Options)
	ClientOption   func(*options.ClientOptions)
	DatabaseOption func(*options.DatabaseOptions)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithOtelEnabled(v bool) Option {
	return func(o *Options) {
		o.OtelEnabled = v
	}
}

func WithOtelOptions(v ...otelmongo.Option) Option {
	return func(o *Options) {
		o.OtelOptions = append(o.OtelOptions, v...)
	}
}

func WithClientOptions(v ...ClientOption) Option {
	return func(o *Options) {
		o.ClientOptions = append(o.ClientOptions, v...)
	}
}

func WithDatabaseOptions(v ...DatabaseOption) Option {
	return func(o *Options) {
		o.DatabaseOptions = append(o.DatabaseOptions, v...)
	}
}

func DefaultOptions() Options {
	return Options{
		OtelEnabled: env.GetBool("OTEL_MONGO_ENABLED", env.GetBool("OTEL_ENABLED", false)),
		OtelOptions: []otelmongo.Option{
			otelmongo.WithCommandAttributeDisabled(env.GetBool("OTEL_MONGO_COMMAND_ATTRIBUTE_DISABLED", false)),
		},
		ClientOptions: []ClientOption{
			func(clientOptions *options.ClientOptions) {
				clientOptions.SetReadConcern(readconcern.Majority())
				clientOptions.SetWriteConcern(writeconcern.Majority())
			},
		},
		DatabaseOptions: nil,
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(ctx context.Context, uri string, opts ...Option) (*Persistor, error) {
	o := DefaultOptions()

	// TODO remove once Database attribute is being exposed
	cs, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uri")
	} else if cs.Database == "" {
		return nil, errors.Errorf("missing database name in uri: %s", uri)
	}

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	// apply client options
	clientOptions := options.Client().ApplyURI(uri)
	for _, opt := range o.ClientOptions {
		opt(clientOptions)
	}

	// apply database options
	databaseOptions := options.Database()
	for _, opt := range o.DatabaseOptions {
		opt(databaseOptions)
	}

	// setup otel
	if o.OtelEnabled {
		clientOptions.SetMonitor(
			otelmongo.NewMonitor(o.OtelOptions...),
		)
	}

	// create connection
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect")
	}

	// test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Persistor{
		client: client,
		db:     client.Database(cs.Database, databaseOptions),
	}, nil
}

func (p Persistor) DB() *mongo.Database {
	return p.db
}

func (p Persistor) Client() *mongo.Client {
	return p.client
}

func (p Persistor) Ping(ctx context.Context) error {
	return p.client.Ping(ctx, nil)
}

func (p Persistor) Collection(name string, opts ...CollectionOption) (*Collection, error) {
	return NewCollection(p.db, name, opts...)
}

// HasCollection checks if the given collection exists
func (p Persistor) HasCollection(ctx context.Context, name string) (bool, error) {
	names, err := p.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}
	for i := range names {
		if names[i] == name {
			return true, nil
		}
	}
	return false, nil
}

func (p Persistor) Close(ctx context.Context) error {
	return p.client.Disconnect(ctx)
}
