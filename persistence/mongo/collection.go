package keelmongo

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	keelerrors "github.com/foomo/keel/errors"
	keelpersistence "github.com/foomo/keel/persistence"
	keeltime "github.com/foomo/keel/time"
)

type (
	DecodeFn         func(val interface{}) error
	IterateHandlerFn func(decode DecodeFn) error
)

// Collection can only be used in the Persistor.WithCollection call.ss
type (
	Collection struct {
		db         *mongo.Database
		collection *mongo.Collection
	}
	CollectionOptions struct {
		*options.CollectionOptions
		*options.CreateIndexesOptions
		Indexes        []mongo.IndexModel
		IndexesContext context.Context
	}
	CollectionOption func(*CollectionOptions)
)

func DefaultCollectionOptions() CollectionOptions {
	return CollectionOptions{
		CollectionOptions:    options.Collection(),
		CreateIndexesOptions: options.CreateIndexes(),
		IndexesContext:       context.Background(),
	}
}

func CollectionWithReadConcern(v *readconcern.ReadConcern) CollectionOption {
	return func(o *CollectionOptions) {
		o.CollectionOptions.SetReadConcern(v)
	}
}

func CollectionWithWriteConcern(v *writeconcern.WriteConcern) CollectionOption {
	return func(o *CollectionOptions) {
		o.CollectionOptions.SetWriteConcern(v)
	}
}

func CollectionWithReadPreference(v *readpref.ReadPref) CollectionOption {
	return func(o *CollectionOptions) {
		o.CollectionOptions.SetReadPreference(v)
	}
}

func CollectionWithRegistry(v *bsoncodec.Registry) CollectionOption {
	return func(o *CollectionOptions) {
		o.CollectionOptions.SetRegistry(v)
	}
}

func CollectionWithIndexes(v ...mongo.IndexModel) CollectionOption {
	return func(o *CollectionOptions) {
		o.Indexes = v
	}
}

func CollectionWithIndexesMaxTime(v time.Duration) CollectionOption {
	return func(o *CollectionOptions) {
		o.CreateIndexesOptions.SetMaxTime(v)
	}
}

func CollectionWithIndexesContext(v int32) CollectionOption {
	return func(o *CollectionOptions) {
		o.CreateIndexesOptions.SetCommitQuorumInt(v)
	}
}

func CollectionWithIndexesQuorumMajority() CollectionOption {
	return func(o *CollectionOptions) {
		o.CreateIndexesOptions.SetCommitQuorumMajority()
	}
}

func CollectionWithIndexesCommitQuorumString(v string) CollectionOption {
	return func(o *CollectionOptions) {
		o.CreateIndexesOptions.SetCommitQuorumString(v)
	}
}

func CollectionWithIndexesCommitQuorumVotingMembers(v context.Context) CollectionOption {
	return func(o *CollectionOptions) {
		o.CreateIndexesOptions.SetCommitQuorumVotingMembers()
	}
}

func NewCollection(db *mongo.Database, name string, opts ...CollectionOption) (*Collection, error) {
	o := DefaultCollectionOptions()
	for _, opt := range opts {
		opt(&o)
	}

	col := db.Collection(name, o.CollectionOptions)

	if len(o.Indexes) > 0 {
		if _, err := col.Indexes().CreateMany(o.IndexesContext, o.Indexes, o.CreateIndexesOptions); err != nil {
			return nil, err
		}
	}

	return &Collection{
		db:         db,
		collection: col,
	}, nil
}

func (c *Collection) DB() *mongo.Database {
	return c.db
}

func (c *Collection) Col() *mongo.Collection {
	return c.collection
}

// Get ...
func (c *Collection) Get(ctx context.Context, id string, result interface{}, opts ...*options.FindOneOptions) error {
	if id == "" {
		return keelpersistence.ErrNotFound
	}
	return c.FindOne(ctx, bson.M{"id": id}, result, opts...)
}

// Exists ...
func (c *Collection) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}
	ret, err := c.collection.CountDocuments(ctx, bson.M{"id": id})
	return ret > 0, err
}

// Upsert inserts/updates with protection against dirty-writes
// requires an unique index on id AND id + version
func (c *Collection) Upsert(ctx context.Context, id string, entity Entity) error {
	if id == "" {
		return errors.New("id must not be empty")
	} else if entity == nil {
		return errors.New("entity must not be nil")
	}

	if v, ok := entity.(EntityWithTimestamps); ok {
		now := keeltime.Now()
		if ct := v.GetCreatedAt(); ct.IsZero() {
			v.SetCreatedAt(now)
		}
		v.SetUpdatedAt(now)
	}

	if v, ok := entity.(EntityWithVersion); ok {
		currentVersion := v.GetVersion()
		// increment version
		v.IncreaseVersion()

		if currentVersion == 0 {
			// insert the new document
			return c.Insert(ctx, entity)
		} else if err := c.collection.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "id", Value: id}, {Key: "version", Value: currentVersion}},
			bson.D{{Key: "$set", Value: entity}},
			options.FindOneAndUpdate().SetUpsert(false),
		).Err(); errors.Is(err, mongo.ErrNoDocuments) {
			return keelerrors.NewWrappedError(keelpersistence.ErrDirtyWrite, err)
		} else if err != nil {
			return err
		}
	} else if _, err := c.collection.UpdateOne(
		ctx,
		bson.D{{Key: "id", Value: id}},
		bson.D{{Key: "$set", Value: entity}},
		options.Update().SetUpsert(true),
	); err != nil {
		return err
	}

	return nil
}

func (c *Collection) Insert(ctx context.Context, entity Entity) error {
	if _, err := c.collection.InsertOne(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (c *Collection) Delete(ctx context.Context, id string) error {
	if id == "" {
		return keelpersistence.ErrNotFound
	}
	if err := c.collection.FindOneAndDelete(ctx, bson.M{"id": id}).Err(); errors.Is(err, mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}
	return nil
}

// Find ...
func (c *Collection) Find(ctx context.Context, filter, results interface{}, opts ...*options.FindOptions) error {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}
	defer CloseCursor(ctx, cursor, &err)
	err = cursor.All(ctx, results)
	return err
}

// FindOne ...
func (c *Collection) FindOne(ctx context.Context, filter, result interface{}, opts ...*options.FindOneOptions) error {
	res := c.collection.FindOne(ctx, filter, opts...)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, res.Err())
	} else if res.Err() != nil {
		return res.Err()
	}
	return res.Decode(result)
}

// FindIterate ...
func (c *Collection) FindIterate(ctx context.Context, filter interface{}, handler IterateHandlerFn, opts ...*options.FindOptions) error {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}
	defer CloseCursor(ctx, cursor, &err)
	for cursor.Next(ctx) {
		if err := handler(cursor.Decode); err != nil {
			return err
		}
	}
	return err
}

// Aggregate ...
func (c *Collection) Aggregate(ctx context.Context, pipeline mongo.Pipeline, results interface{}, opts ...*options.AggregateOptions) error {
	cursor, err := c.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer CloseCursor(ctx, cursor, &err)
	err = cursor.All(ctx, results)
	return err
}

func (c *Collection) AggregateIterate(
	ctx context.Context,
	pipeline mongo.Pipeline,
	handler IterateHandlerFn,
	opts ...*options.AggregateOptions,
) error {
	cursor, err := c.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer CloseCursor(ctx, cursor, &err)
	for cursor.Next(ctx) {
		if err := handler(cursor.Decode); err != nil {
			return err
		}
	}
	return nil
}

// Count returns the count of documents
func (c *Collection) Count(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return c.collection.CountDocuments(ctx, filter, opts...)
}

// CountAll returns the count of all documents
func (c *Collection) CountAll(ctx context.Context) (int64, error) {
	return c.Count(ctx, bson.D{})
}
