package keelmongo

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

// Persistor exported to used also for embedding into other types in foreign packages.
type (
	Persistor struct {
		client *mongo.Client
		db     *mongo.Database
	}
	Options struct {
		OtelEnabled     bool
		OtelServiceName string
		OtelOptions     []otelmongo.Option
		ClientOptions   *options.ClientOptions
		DatabaseOptions *options.DatabaseOptions
	}
	Option func(o *Options)
)

func WithOtelEnabled(v bool) Option {
	return func(o *Options) {
		o.OtelEnabled = v
	}
}

func WithOtelServiceName(v string) Option {
	return func(o *Options) {
		o.OtelServiceName = v
	}
}

func WithOtelOptions(v ...otelmongo.Option) Option {
	return func(o *Options) {
		o.OtelOptions = append(o.OtelOptions, v...)
	}
}

func WithClientOptions(v *options.ClientOptions) Option {
	return func(o *Options) {
		o.ClientOptions = options.MergeClientOptions(o.ClientOptions, v)
	}
}

func WithDatabaseOptions(v *options.DatabaseOptions) Option {
	return func(o *Options) {
		o.DatabaseOptions = options.MergeDatabaseOptions(o.DatabaseOptions, v)
	}
}

func DefaultOptions() Options {
	return Options{
		OtelEnabled:     true,
		OtelServiceName: "mongo",
		OtelOptions:     nil,
		ClientOptions: options.Client().
			SetReadConcern(readconcern.Majority()).
			SetReadPreference(readpref.Nearest()).
			SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
		DatabaseOptions: nil,
	}
}

// New ...
func New(ctx context.Context, uri string, opts ...Option) (*Persistor, error) {
	o := DefaultOptions()

	// TODO remove once Database attribute is being exposed
	cs, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uri")
	} else if cs.Database == "" {
		return nil, errors.Errorf("missing database name in uri: %s", uri)
	}

	// apply uri
	o.ClientOptions.ApplyURI(uri)

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	// setup otel
	if o.OtelEnabled {
		o.ClientOptions.SetMonitor(otelmongo.NewMonitor(o.OtelServiceName, o.OtelOptions...))
	}

	// create connection
	client, err := mongo.Connect(ctx, o.ClientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect")
	}

	// test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Persistor{
		client: client,
		db:     client.Database(cs.Database),
	}, nil
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
