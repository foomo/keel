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
// named index ever created via NewCollection. It exists purely so Readme()
// can render the docs table — nothing in the runtime path reads it.
// registryMu guards both maps: NewCollection writes them and Readme reads
// them, potentially from different goroutines.
var (
	dbs        = map[string][]string{}
	indices    = map[string]map[string][]string{}
	registryMu sync.RWMutex
)

type (
	DecodeFn         func(val any) error
	IterateHandlerFn func(decode DecodeFn) error
)

// Collection can only be used in the Persistor.WithCollection call.ss
type (
	Collection struct {
		db         *mongo.Database
		collection *mongo.Collection
	}
	CollectionOptions struct {
		*options.CollectionOptionsBuilder
		*options.CreateIndexesOptionsBuilder
		Indexes        []mongo.IndexModel
		IndexesMaxTime time.Duration
	}
	CollectionOption func(*CollectionOptions)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func DefaultCollectionOptions() CollectionOptions {
	return CollectionOptions{
		CollectionOptionsBuilder:    options.Collection(),
		CreateIndexesOptionsBuilder: options.CreateIndexes(),
	}
}

func CollectionWithReadConcern(v *readconcern.ReadConcern) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetReadConcern(v)
	}
}

func CollectionWithWriteConcern(v *writeconcern.WriteConcern) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetWriteConcern(v)
	}
}

func CollectionWithReadPreference(v *readpref.ReadPref) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetReadPreference(v)
	}
}

func CollectionWithRegistry(v *bson.Registry) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetRegistry(v)
	}
}

func CollectionWithIndexes(v ...mongo.IndexModel) CollectionOption {
	return func(o *CollectionOptions) {
		o.Indexes = v
	}
}

func CollectionWithIndexesMaxTime(v time.Duration) CollectionOption {
	return func(o *CollectionOptions) {
		o.IndexesMaxTime = v
	}
}

func CollectionWithCommitQuorumInt(v int32) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumInt(v)
	}
}

func CollectionWithCommitQuorumMajority() CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumMajority()
	}
}

func CollectionWithCommitQuorumString(v string) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumString(v)
	}
}

func CollectionWithCommitQuorumVotingMembers(v context.Context) CollectionOption {
	return func(o *CollectionOptions) {
		o.SetCommitQuorumVotingMembers()
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

func (c *Collection) DB() *mongo.Database {
	return c.db
}

func (c *Collection) Col() *mongo.Collection {
	return c.collection
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Collection) Get(ctx context.Context, id string, result any, opts ...options.Lister[options.FindOneOptions]) error {
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

// UpsertUpdate builds the update document for an upsert of entity: the entity
// itself under $set, plus any fields it declares via EntityWithInsertOnlyFields
// under $setOnInsert. An upsert that resolves to an update therefore leaves those
// fields at their stored value, while one that inserts writes them once.
//
// Upsert and UpsertMany use this internally. It is exported so callers driving
// their own bulk writes - an import feed filtering on a staleness predicate, say -
// can build the same document without restating the policy. Such callers remain
// responsible for the filter, for stamping updatedAt (see EntityWithTimestamps),
// and for executing the write.
//
// The $setOnInsert clause only takes effect when the write model enables upserting;
// with SetUpsert(false) it is inert.
func UpsertUpdate(entity Entity, now time.Time) bson.D {
	update := bson.D{bson.E{Key: "$set", Value: entity}}

	if v, ok := entity.(EntityWithInsertOnlyFields); ok {
		if insertOnly := v.InsertOnlyFields(now); len(insertOnly) > 0 {
			update = append(update, bson.E{Key: "$setOnInsert", Value: insertOnly})
		}
	}

	return update
}

// ApplyTimestamps stamps updatedAt on entity, and fills in a missing createdAt
// unless the entity defers that to $setOnInsert via EntityWithInsertOnlyFields.
// Entities that do not implement EntityWithTimestamps are left untouched.
//
// Upsert and UpsertMany call this for you. It is exported as the companion to
// UpsertUpdate, so callers driving their own bulk writes apply the same timestamp
// policy rather than restating it per call site - in particular, a blind writer
// must not stamp createdAt itself, or it overwrites the stored value.
//
// Note that this mutates entity.
func ApplyTimestamps(entity Entity, now time.Time) {
	v, ok := entity.(EntityWithTimestamps)
	if !ok {
		return
	}

	// An entity declaring insert-only fields carries its own createdAt policy, so
	// leave a zero value alone: $setOnInsert supplies it exactly when the write
	// turns out to be an insert, and omitempty keeps it out of $set otherwise.
	_, insertOnly := entity.(EntityWithInsertOnlyFields)
	if ct := v.GetCreatedAt(); ct.IsZero() && !insertOnly {
		v.SetCreatedAt(now)
	}

	v.SetUpdatedAt(now)
}

// stampCreatedAt fills in a missing createdAt on the paths that insert the
// document directly, where $setOnInsert never applies.
func stampCreatedAt(entity Entity, now time.Time) {
	if v, ok := entity.(EntityWithTimestamps); ok && v.GetCreatedAt().IsZero() {
		v.SetCreatedAt(now)
	}
}

func (c *Collection) Upsert(ctx context.Context, id string, entity Entity) error {
	if id == "" {
		return errors.New("id must not be empty")
	} else if entity == nil {
		return errors.New("entity must not be nil")
	}

	now := time.Now()

	ApplyTimestamps(entity, now)

	if v, ok := entity.(EntityWithVersion); ok {
		currentVersion := v.GetVersion()
		// increment version
		v.IncreaseVersion()

		if currentVersion == 0 {
			// insert the new document
			stampCreatedAt(entity, now)

			return c.Insert(ctx, entity)
		} else if err := c.collection.FindOneAndUpdate(
			ctx,
			bson.D{bson.E{Key: "id", Value: id}, bson.E{Key: "version", Value: currentVersion}},
			// never upserts, so insert-only fields cannot apply here
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
		UpsertUpdate(entity, now),
		options.UpdateOne().SetUpsert(true),
	); err != nil {
		return err
	}

	return nil
}

// UpsertMany - NOTE: upsert many does NOT through an explicit error on dirty write so we can only assume it.
func (c *Collection) UpsertMany(ctx context.Context, entities []Entity) error {
	var (
		versionUpserts int64
		operations     []mongo.WriteModel
	)

	// one timestamp for the whole batch, so entities written together agree
	now := time.Now()

	for _, entity := range entities {
		if entity == nil {
			return errors.New("entity must not be nil")
		} else if entity.GetID() == "" {
			return errors.New("id must not be empty")
		}

		ApplyTimestamps(entity, now)

		if v, ok := entity.(EntityWithVersion); ok {
			currentVersion := v.GetVersion()
			// increment version
			v.IncreaseVersion()

			if currentVersion == 0 {
				stampCreatedAt(entity, now)

				operations = append(operations,
					mongo.NewInsertOneModel().SetDocument(entity),
				)
			} else {
				versionUpserts++

				operations = append(operations,
					mongo.NewUpdateOneModel().
						SetFilter(bson.D{bson.E{Key: "id", Value: entity.GetID()}, bson.E{Key: "version", Value: currentVersion}}).
						// never upserts, so insert-only fields cannot apply here
						SetUpdate(bson.D{bson.E{Key: "$set", Value: entity}}).
						SetUpsert(false),
				)
			}
		} else {
			operations = append(operations,
				mongo.NewUpdateOneModel().
					SetFilter(bson.D{bson.E{Key: "id", Value: entity.GetID()}}).
					SetUpdate(UpsertUpdate(entity, now)).
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

func (c *Collection) FindOne(ctx context.Context, filter, result any, opts ...options.Lister[options.FindOneOptions]) error {
	res := c.collection.FindOne(ctx, filter, opts...)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return errors.Join(keelpersistence.ErrNotFound, res.Err())
	} else if res.Err() != nil {
		return res.Err()
	}

	return res.Decode(result)
}

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

// Count returns the count of documents
func (c *Collection) Count(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return c.collection.CountDocuments(ctx, filter, opts...)
}

// CountAll returns the count of all documents
func (c *Collection) CountAll(ctx context.Context) (int64, error) {
	return c.Count(ctx, bson.D{})
}
