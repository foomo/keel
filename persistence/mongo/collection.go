package keelmongo

import (
	"context"
	"slices"
	"time"

	keelerrors "github.com/foomo/keel/errors"
	keelpersistence "github.com/foomo/keel/persistence"
	keeltime "github.com/foomo/keel/time"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

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

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCollection(db *mongo.Database, name string, opts ...CollectionOption) (*Collection, error) {
	o := DefaultCollectionOptions()
	for _, opt := range opts {
		opt(&o)
	}

	col := db.Collection(name, o.CollectionOptions)
	if !slices.Contains(dbs[db.Name()], name) {
		dbs[db.Name()] = append(dbs[db.Name()], name)
	}

	if len(o.Indexes) > 0 {
		if _, err := col.Indexes().CreateMany(o.IndexesContext, o.Indexes, o.CreateIndexesOptions); err != nil {
			return nil, err
		}
		if _, ok := indices[db.Name()]; !ok {
			indices[db.Name()] = map[string][]string{}
		}
		for _, index := range o.Indexes {
			if index.Options != nil && index.Options.Name != nil {
				indices[db.Name()][name] = append(indices[db.Name()][name], *index.Options.Name)
			}
		}
	}

	return &Collection{
		db:         db,
		collection: col,
	}, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

func (c *Collection) DB() *mongo.Database {
	return c.db
}

func (c *Collection) Col() *mongo.Collection {
	return c.collection
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Collection) Get(ctx context.Context, id string, result interface{}, opts ...*options.FindOneOptions) error {
	if id == "" {
		return keelpersistence.ErrNotFound
	}
	return c.FindOne(ctx, bson.M{"id": id}, result, opts...)
}

func (c *Collection) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}
	ret, err := c.collection.CountDocuments(ctx, bson.M{"id": id})
	return ret > 0, err
}

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
			bson.D{bson.E{Key: "id", Value: id}, bson.E{Key: "version", Value: currentVersion}},
			bson.D{bson.E{Key: "$set", Value: entity}},
			options.FindOneAndUpdate().SetUpsert(false),
		).Err(); errors.Is(err, mongo.ErrNoDocuments) {
			return keelerrors.NewWrappedError(keelpersistence.ErrDirtyWrite, err)
		} else if err != nil {
			return err
		}
	} else if _, err := c.collection.UpdateOne(
		ctx,
		bson.D{bson.E{Key: "id", Value: id}},
		bson.D{bson.E{Key: "$set", Value: entity}},
		options.Update().SetUpsert(true),
	); err != nil {
		return err
	}

	return nil
}

// UpsertMany - NOTE: upsert many does NOT through an explicit error on dirty write so we can only assume it.
func (c *Collection) UpsertMany(ctx context.Context, entities []Entity) error {
	var versionUpserts int64
	var operations []mongo.WriteModel

	for _, entity := range entities {
		if entity == nil {
			return errors.New("entity must not be nil")
		} else if entity.GetID() == "" {
			return errors.New("id must not be empty")
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
				operations = append(operations,
					mongo.NewInsertOneModel().SetDocument(entity),
				)
			} else {
				versionUpserts++
				operations = append(operations,
					mongo.NewUpdateOneModel().
						SetFilter(bson.D{bson.E{Key: "id", Value: entity.GetID()}, bson.E{Key: "version", Value: currentVersion}}).
						SetUpdate(bson.D{bson.E{Key: "$set", Value: entity}}).
						SetUpsert(false),
				)
			}
		} else {
			operations = append(operations,
				mongo.NewUpdateOneModel().
					SetFilter(bson.D{bson.E{Key: "id", Value: entity.GetID()}}).
					SetUpdate(bson.D{bson.E{Key: "$set", Value: entity}}).
					SetUpsert(true),
			)
		}
	}

	// Specify an option to turn the bulk insertion in order of operation
	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(false)

	res, err := c.Col().BulkWrite(ctx, operations, &bulkOption)
	if err != nil {
		return err
	} else if versionUpserts > 0 && (res.MatchedCount < versionUpserts || res.ModifiedCount != res.MatchedCount) {
		// log.Logger().Info("missing upserts",
		// 	zap.Int64("MatchedCount", res.MatchedCount),
		// 	zap.Int64("InsertedCount", res.InsertedCount),
		// 	zap.Int64("UpsertedCount", res.UpsertedCount),
		// 	zap.Int64("ModifiedCount", res.ModifiedCount),
		// 	zap.Any("UpsertedIDs", res.UpsertedIDs),
		// 	zap.Any("versionUpserts", versionUpserts),
		// )
		return keelpersistence.ErrDirtyWrite
	}

	return nil
}

func (c *Collection) Insert(ctx context.Context, entity Entity) error {
	if entity == nil {
		return errors.New("entity must not be nil")
	} else if entity.GetID() == "" {
		return errors.New("id must not be empty")
	}

	if v, ok := entity.(EntityWithTimestamps); ok {
		now := keeltime.Now()
		if ct := v.GetCreatedAt(); ct.IsZero() {
			v.SetCreatedAt(now)
		}
		v.SetUpdatedAt(now)
	}

	if v, ok := entity.(EntityWithVersion); ok {
		// increment version
		v.IncreaseVersion()
	}

	if _, err := c.collection.InsertOne(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (c *Collection) InsertMany(ctx context.Context, entities []Entity) error {
	inserts := make([]interface{}, len(entities))
	for i, entity := range entities {
		if entity == nil {
			return errors.New("entity must not be nil")
		} else if entity.GetID() == "" {
			return errors.New("id must not be empty")
		}

		if v, ok := entity.(EntityWithTimestamps); ok {
			now := keeltime.Now()
			if ct := v.GetCreatedAt(); ct.IsZero() {
				v.SetCreatedAt(now)
			}
			v.SetUpdatedAt(now)
		}

		if v, ok := entity.(EntityWithVersion); ok {
			// increment version
			v.IncreaseVersion()
		}

		inserts[i] = entity
	}

	if _, err := c.collection.InsertMany(ctx, inserts); err != nil {
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

func (c *Collection) Find(ctx context.Context, filter, results interface{}, opts ...*options.FindOptions) error {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}

	if err = cursor.All(ctx, results); err != nil {
		return err
	}

	return cursor.Err()
}

func (c *Collection) FindOne(ctx context.Context, filter, result interface{}, opts ...*options.FindOneOptions) error {
	res := c.collection.FindOne(ctx, filter, opts...)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, res.Err())
	} else if res.Err() != nil {
		return res.Err()
	}

	return res.Decode(result)
}

func (c *Collection) FindIterate(ctx context.Context, filter interface{}, handler IterateHandlerFn, opts ...*options.FindOptions) error {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return keelerrors.NewWrappedError(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}

	defer CloseCursor(context.WithoutCancel(ctx), cursor)

	for cursor.Next(ctx) {
		if err := handler(cursor.Decode); err != nil {
			return err
		}
	}

	return cursor.Err()
}

func (c *Collection) Aggregate(ctx context.Context, pipeline mongo.Pipeline, results interface{}, opts ...*options.AggregateOptions) error {
	cursor, err := c.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}

	if err = cursor.All(ctx, results); err != nil {
		return err
	}

	return cursor.Err()
}

func (c *Collection) AggregateIterate(ctx context.Context, pipeline mongo.Pipeline, handler IterateHandlerFn, opts ...*options.AggregateOptions) error {
	cursor, err := c.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}

	defer CloseCursor(context.WithoutCancel(ctx), cursor)

	for cursor.Next(ctx) {
		if err := handler(cursor.Decode); err != nil {
			return err
		}
	}

	return cursor.Err()
}

// Count returns the count of documents
func (c *Collection) Count(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return c.collection.CountDocuments(ctx, filter, opts...)
}

// CountAll returns the count of all documents
func (c *Collection) CountAll(ctx context.Context) (int64, error) {
	return c.Count(ctx, bson.D{})
}
