package keelmongo

import (
	"context"
	"slices"

	"github.com/foomo/keel/env"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
	"go.uber.org/zap"
)

type (
	// Persistor holds a MongoDB client and the database named in its
	// connection URI. It is exported so it can be embedded into types in
	// other packages.
	Persistor struct {
		client *mongo.Client
		db     *mongo.Database
	}
	// Options configures [New].
	Options struct {
		// OtelEnabled enables the OpenTelemetry command monitor.
		OtelEnabled bool
		// OtelOptions configure the OpenTelemetry command monitor.
		OtelOptions []otelmongo.Option
		// ClientOptions are applied to the client options after the URI.
		ClientOptions []ClientOption
		// ClientLoggerOptions configure the driver logger. They are only
		// applied if no logger options were set via ClientOptions.
		ClientLoggerOptions []ClientLoggerOption
		// DatabaseOptions are applied to the database options.
		DatabaseOptions []DatabaseOption
	}
	// Option configures [Options].
	Option func(o *Options)
	// ClientOption configures the driver's client options.
	ClientOption func(*options.ClientOptions)
	// ClientLoggerOption configures the driver's logger options.
	ClientLoggerOption func(*options.LoggerOptions)
	// DatabaseOption configures the driver's database options.
	DatabaseOption func(*options.DatabaseOptionsBuilder)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithOtelEnabled sets whether OpenTelemetry instrumentation is enabled.
// Defaults to the OTEL_MONGO_ENABLED environment variable, falling back to
// OTEL_ENABLED, or false.
func WithOtelEnabled(v bool) Option {
	return func(o *Options) {
		o.OtelEnabled = v
	}
}

// WithOtelOptions appends options for the OpenTelemetry command monitor.
func WithOtelOptions(v ...otelmongo.Option) Option {
	return func(o *Options) {
		o.OtelOptions = append(o.OtelOptions, v...)
	}
}

// WithClientOptions appends client options.
func WithClientOptions(v ...ClientOption) Option {
	return func(o *Options) {
		o.ClientOptions = append(o.ClientOptions, v...)
	}
}

// WithClientLogger routes the driver's log output to the given zap logger.
func WithClientLogger(v *zap.Logger) Option {
	return func(o *Options) {
		o.ClientLoggerOptions = append(o.ClientLoggerOptions, func(o *options.LoggerOptions) {
			o.SetSink(zapr.NewLogger(v).GetSink())
		})
	}
}

// WithClientLoggerComponentLevel sets the driver log level for component c.
func WithClientLoggerComponentLevel(c options.LogComponent, l options.LogLevel) Option {
	return func(o *Options) {
		o.ClientLoggerOptions = append(o.ClientLoggerOptions, func(o *options.LoggerOptions) {
			o.SetComponentLevel(c, l)
		})
	}
}

// WithClientCompression is an [Option] enabling snappy and zstd wire
// compression.
func WithClientCompression(o *Options) {
	o.ClientOptions = append(o.ClientOptions, func(o *options.ClientOptions) {
		o.SetCompressors([]string{"snappy", "zstd"})
	})
}

// WithDatabaseOptions appends database options.
func WithDatabaseOptions(v ...DatabaseOption) Option {
	return func(o *Options) {
		o.DatabaseOptions = append(o.DatabaseOptions, v...)
	}
}

// DefaultOptions returns the default [Options]: majority read and write
// concern, OpenTelemetry enabled per OTEL_MONGO_ENABLED or OTEL_ENABLED, and
// command attributes disabled per OTEL_MONGO_COMMAND_ATTRIBUTE_DISABLED.
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
		ClientLoggerOptions: nil,
		DatabaseOptions:     nil,
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New connects to MongoDB using uri, pings the server and returns a
// [Persistor] for the database named in uri. It returns an error if uri is
// invalid, names no database, or the connection or ping fails.
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

	if clientOptions.LoggerOptions == nil && len(o.ClientLoggerOptions) > 0 {
		clientOptions.LoggerOptions = options.Logger()
		for _, opt := range o.ClientLoggerOptions {
			opt(clientOptions.LoggerOptions)
		}
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
	client, err := mongo.Connect(clientOptions)
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

// DB returns the database.
func (p Persistor) DB() *mongo.Database {
	return p.db
}

// Client returns the underlying [mongo.Client].
func (p Persistor) Client() *mongo.Client {
	return p.client
}

// Ping verifies the connection to the server.
func (p Persistor) Ping(ctx context.Context) error {
	return p.client.Ping(ctx, nil)
}

// Collection returns the named [Collection] via [NewCollection].
func (p Persistor) Collection(name string, opts ...CollectionOption) (*Collection, error) {
	return NewCollection(p.db, name, opts...)
}

// HasCollection reports whether a collection with the given name exists.
func (p Persistor) HasCollection(ctx context.Context, name string) (bool, error) {
	names, err := p.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}

	if slices.Contains(names, name) {
		return true, nil
	}

	return false, nil
}

// Close disconnects the client.
func (p Persistor) Close(ctx context.Context) error {
	return p.client.Disconnect(ctx)
}
