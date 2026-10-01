package keelmongo

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	keelpersistence "github.com/foomo/keel/persistence"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// dbs and indices form a package-level registry of every collection and
// named index ever created via [NewCollection]. It exists purely so [Readme]
// can render the docs table — nothing in the runtime path reads it.
// registryMu guards both maps: NewCollection writes them and Readme reads
// them, potentially from different goroutines.
var (
	dbs        = map[string][]string{}
	indices    = map[string]map[string][]string{}
	registryMu sync.RWMutex
)

type (
	// DecodeFn decodes the current document into val.
	DecodeFn func(val any) error
	// IterateHandlerFn is called once per document by the iterating
	// methods; returning an error stops the iteration and is returned to
	// the caller.
	IterateHandlerFn func(decode DecodeFn) error
)

type (
	// Collection wraps a [mongo.Collection] whose documents are keyed by an
	// "id" field. Create it with [NewCollection] or [Persistor.Collection].
	Collection struct {
		db         *mongo.Database
		collection *mongo.Collection
	}
	// CollectionOptions configures [NewCollection].
	CollectionOptions struct {
		*options.CollectionOptionsBuilder
		*options.CreateIndexesOptionsBuilder
		// Indexes are created when the collection is set up.
		Indexes []mongo.IndexModel
		// IndexesMaxTime bounds index creation; zero means no timeout.
		IndexesMaxTime time.Duration
	}
	// CollectionOption configures [CollectionOptions].
	CollectionOption func(*CollectionOptions)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// DefaultCollectionOptions returns empty collection and index creation
// options.
func DefaultCollectionOptions() CollectionOptions {
	return CollectionOptions{
		CollectionOptionsBuilder:    options.Collection(),
		CreateIndexesOptionsBuilder: options.CreateIndexes(),
	}
}

// CollectionWithReadConcern sets the collection's read concern.
func CollectionWithReadConcern(v *readconcern.ReadConcern) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetReadConcern(v)
	}
}

// CollectionWithWriteConcern sets the collection's write concern.
func CollectionWithWriteConcern(v *writeconcern.WriteConcern) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetWriteConcern(v)
	}
}

// CollectionWithReadPreference sets the collection's read preference.
func CollectionWithReadPreference(v *readpref.ReadPref) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetReadPreference(v)
	}
}

// CollectionWithRegistry sets the BSON registry used by the collection.
func CollectionWithRegistry(v *bson.Registry) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetRegistry(v)
	}
}

// CollectionWithIndexes sets the indexes created by [NewCollection],
// replacing any previously set.
func CollectionWithIndexes(v ...mongo.IndexModel) CollectionOption {
	return func(o *CollectionOptions) {
		o.Indexes = v
	}
}

// CollectionWithIndexesMaxTime sets the timeout for index creation.
// Defaults to no timeout.
func CollectionWithIndexesMaxTime(v time.Duration) CollectionOption {
	return func(o *CollectionOptions) {
		o.IndexesMaxTime = v
	}
}

// CollectionWithCommitQuorumInt sets the index build commit quorum to the
// given number of data-bearing voting members.
func CollectionWithCommitQuorumInt(v int32) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumInt(v)
	}
}

// CollectionWithCommitQuorumMajority sets the index build commit quorum to
// "majority".
func CollectionWithCommitQuorumMajority() CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumMajority()
	}
}

// CollectionWithCommitQuorumString sets the index build commit quorum to the
// given string, e.g. a replica set tag name.
func CollectionWithCommitQuorumString(v string) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumString(v)
	}
}

// CollectionWithCommitQuorumVotingMembers sets the index build commit quorum
// to "votingMembers". The context argument is unused.
func CollectionWithCommitQuorumVotingMembers(v context.Context) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumVotingMembers()
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewCollection returns a [Collection] for name in db, registers it for
// [Readme] and synchronously creates any configured indexes. It returns an
// error if index creation fails. It is safe for concurrent use.
func NewCollection(db *mongo.Database, name string, opts ...CollectionOption) (*Collection, error) {
	o := DefaultCollectionOptions()
	for _, opt := range opts {
		opt(&o)
	}

	col := db.Collection(name, o.CollectionOptionsBuilder)

	// Take a write lock for the rest of the function: both dbs and indices
	// are package-level maps and we may mutate either (or both) below.
	registryMu.Lock()
	defer registryMu.Unlock()

	if !slices.Contains(dbs[db.Name()], name) {
		dbs[db.Name()] = append(dbs[db.Name()], name)
	}

	if len(o.Indexes) > 0 {
		if err := func(ctx context.Context) error {
			if o.IndexesMaxTime > 0 {
				var cancel context.CancelFunc

				ctx, cancel = context.WithTimeout(ctx, o.IndexesMaxTime)
				defer cancel()
			}

			if _, err := col.Indexes().CreateMany(ctx, o.Indexes, o.CreateIndexesOptionsBuilder); err != nil {
				return err
			}

			if _, ok := indices[db.Name()]; !ok {
				indices[db.Name()] = map[string][]string{}
			}

			for _, index := range o.Indexes {
				if index.Options != nil {
					var indexOpts options.IndexOptions
					for _, set := range index.Options.Opts {
						_ = set(&indexOpts)
					}

					if indexOpts.Name != nil {
						indices[db.Name()][name] = append(indices[db.Name()][name], *indexOpts.Name)
					}
				}
			}

			return nil
		}(context.Background()); err != nil {
			return nil, err
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

// DB returns the database the collection belongs to.
func (c *Collection) DB() *mongo.Database {
	return c.db
}

// Col returns the underlying [mongo.Collection].
func (c *Collection) Col() *mongo.Collection {
	return c.collection
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Get decodes the document with the given id into result. It returns
// [keelpersistence.ErrNotFound] if id is empty or no document matches.
func (c *Collection) Get(ctx context.Context, id string, result any, opts ...options.Lister[options.FindOneOptions]) error {
	if id == "" {
		return keelpersistence.ErrNotFound
	}

	return c.FindOne(ctx, bson.M{"id": id}, result, opts...)
}

// Exists reports whether a document with the given id exists. An empty id
// reports false.
func (c *Collection) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}

	ret, err := c.collection.CountDocuments(ctx, bson.M{"id": id})

	return ret > 0, err
}

// Upsert stores entity under id. Timestamps are maintained for
// [EntityWithTimestamps]. For [EntityWithVersion] the version is incremented;
// a version-zero entity is inserted via [Collection.Insert], otherwise the
// update only matches the previous version and fails with
// [keelpersistence.ErrDirtyWrite] if none matches. Other entities are
// upserted by id. It returns an error if id is empty or entity is nil.
func (c *Collection) Upsert(ctx context.Context, id string, entity Entity) error {
	if id == "" {
		return errors.New("id must not be empty")
	} else if entity == nil {
		return errors.New("entity must not be nil")
	}

	if v, ok := entity.(EntityWithTimestamps); ok {
		now := time.Now()
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
			return errors.Join(keelpersistence.ErrDirtyWrite, err)
		} else if err != nil {
			return err
		}
	} else if _, err := c.collection.UpdateOne(
		ctx,
		bson.D{bson.E{Key: "id", Value: id}},
		bson.D{bson.E{Key: "$set", Value: entity}},
		options.UpdateOne().SetUpsert(true),
	); err != nil {
		return err
	}

	return nil
}

// UpsertMany stores entities in a single unordered bulk write, applying the
// same timestamp and version rules as [Collection.Upsert] using each
// entity's ID. The bulk write does not report dirty writes per document, so
// [keelpersistence.ErrDirtyWrite] is returned when fewer versioned documents
// matched or were modified than expected. It returns an error if any entity
// is nil or has an empty ID.
func (c *Collection) UpsertMany(ctx context.Context, entities []Entity) error {
	var (
		versionUpserts int64
		operations     []mongo.WriteModel
	)

	for _, entity := range entities {
		if entity == nil {
			return errors.New("entity must not be nil")
		} else if entity.GetID() == "" {
			return errors.New("id must not be empty")
		}

		if v, ok := entity.(EntityWithTimestamps); ok {
			now := time.Now()
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

	res, err := c.Col().BulkWrite(ctx, operations, options.BulkWrite().SetOrdered(false))
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

// Insert inserts entity, setting timestamps for [EntityWithTimestamps] and
// incrementing the version for [EntityWithVersion]. It returns an error if
// entity is nil or has an empty ID.
func (c *Collection) Insert(ctx context.Context, entity Entity) error {
	if entity == nil {
		return errors.New("entity must not be nil")
	} else if entity.GetID() == "" {
		return errors.New("id must not be empty")
	}

	if v, ok := entity.(EntityWithTimestamps); ok {
		now := time.Now()
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

// InsertMany inserts entities, applying the same rules as [Collection.Insert]
// to each. It returns an error if any entity is nil or has an empty ID.
func (c *Collection) InsertMany(ctx context.Context, entities []Entity) error {
	inserts := make([]any, len(entities))
	for i, entity := range entities {
		if entity == nil {
			return errors.New("entity must not be nil")
		} else if entity.GetID() == "" {
			return errors.New("id must not be empty")
		}

		if v, ok := entity.(EntityWithTimestamps); ok {
			now := time.Now()
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

// Delete removes the document with the given id. It returns
// [keelpersistence.ErrNotFound] if id is empty or no document matches.
func (c *Collection) Delete(ctx context.Context, id string) error {
	if id == "" {
		return keelpersistence.ErrNotFound
	}

	if err := c.collection.FindOneAndDelete(ctx, bson.M{"id": id}).Err(); errors.Is(err, mongo.ErrNoDocuments) {
		return errors.Join(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}

	return nil
}

// Find decodes all documents matching filter into results, which must be a
// pointer to a slice.
func (c *Collection) Find(ctx context.Context, filter, results any, opts ...options.Lister[options.FindOptions]) error {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.Join(keelpersistence.ErrNotFound, err)
	} else if err != nil {
		return err
	}

	if err = cursor.All(ctx, results); err != nil {
		return err
	}

	return cursor.Err()
}

// FindOne decodes the first document matching filter into result. It returns
// [keelpersistence.ErrNotFound] if no document matches.
func (c *Collection) FindOne(ctx context.Context, filter, result any, opts ...options.Lister[options.FindOneOptions]) error {
	res := c.collection.FindOne(ctx, filter, opts...)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return errors.Join(keelpersistence.ErrNotFound, res.Err())
	} else if res.Err() != nil {
		return res.Err()
	}

	return res.Decode(result)
}

// FindIterate calls handler for each document matching filter, stopping at
// the first handler error.
func (c *Collection) FindIterate(ctx context.Context, filter any, handler IterateHandlerFn, opts ...options.Lister[options.FindOptions]) error {
	cursor, err := c.collection.Find(ctx, filter, opts...)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.Join(keelpersistence.ErrNotFound, err)
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

// Aggregate runs pipeline and decodes all resulting documents into results,
// which must be a pointer to a slice.
func (c *Collection) Aggregate(ctx context.Context, pipeline mongo.Pipeline, results any, opts ...options.Lister[options.AggregateOptions]) error {
	cursor, err := c.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}

	if err = cursor.All(ctx, results); err != nil {
		return err
	}

	return cursor.Err()
}

// AggregateIterate runs pipeline and calls handler for each resulting
// document, stopping at the first handler error.
func (c *Collection) AggregateIterate(ctx context.Context, pipeline mongo.Pipeline, handler IterateHandlerFn, opts ...options.Lister[options.AggregateOptions]) error {
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

// Count returns the number of documents matching filter.
func (c *Collection) Count(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return c.collection.CountDocuments(ctx, filter, opts...)
}

// CountAll returns the number of documents in the collection.
func (c *Collection) CountAll(ctx context.Context) (int64, error) {
	return c.Count(ctx, bson.D{})
}
