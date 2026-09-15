package keelmongo_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mongodbcontainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	testMongoImage = "mongo:8"
	testDatabase   = "keeltest"
)

// --- test entities -----------------------------------------------------------

// timestamp mirrors how consumers store times as an integer rather than a
// time.Time, which is what makes `omitempty` unambiguous: the zero value is 0.
type timestamp int64

func newTimestamp(t time.Time) timestamp {
	if t.IsZero() {
		return 0
	}

	return timestamp(t.UnixNano())
}

func (ts timestamp) time() time.Time {
	if ts == 0 {
		return time.Time{}
	}

	return time.Unix(0, int64(ts))
}

// insertOnlyEntity implements EntityWithInsertOnlyFields the way a consumer is
// expected to: createdAt is tagged omitempty so a zero value stays out of $set,
// and InsertOnlyFields returns nothing once the entity carries the value itself.
type insertOnlyEntity struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	CreatedAt timestamp `bson:"createdAt,omitempty"`
	UpdatedAt timestamp `bson:"updatedAt"`
}

func (e *insertOnlyEntity) GetID() string            { return e.ID }
func (e *insertOnlyEntity) SetID(id string)          { e.ID = id }
func (e *insertOnlyEntity) GetCreatedAt() time.Time  { return e.CreatedAt.time() }
func (e *insertOnlyEntity) SetCreatedAt(t time.Time) { e.CreatedAt = newTimestamp(t) }
func (e *insertOnlyEntity) GetUpdatedAt() time.Time  { return e.UpdatedAt.time() }
func (e *insertOnlyEntity) SetUpdatedAt(t time.Time) { e.UpdatedAt = newTimestamp(t) }

func (e *insertOnlyEntity) InsertOnlyFields(now time.Time) bson.D {
	if e.CreatedAt != 0 {
		// already carried in $set; returning it here would collide
		return nil
	}

	return bson.D{{Key: "createdAt", Value: newTimestamp(now)}}
}

// legacyEntity deliberately does NOT implement EntityWithInsertOnlyFields, so it
// pins the pre-existing behaviour every other keel consumer still relies on.
type legacyEntity struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	CreatedAt timestamp `bson:"createdAt"`
	UpdatedAt timestamp `bson:"updatedAt"`
}

func (e *legacyEntity) GetID() string            { return e.ID }
func (e *legacyEntity) SetID(id string)          { e.ID = id }
func (e *legacyEntity) GetCreatedAt() time.Time  { return e.CreatedAt.time() }
func (e *legacyEntity) SetCreatedAt(t time.Time) { e.CreatedAt = newTimestamp(t) }
func (e *legacyEntity) GetUpdatedAt() time.Time  { return e.UpdatedAt.time() }
func (e *legacyEntity) SetUpdatedAt(t time.Time) { e.UpdatedAt = newTimestamp(t) }

// --- harness -----------------------------------------------------------------

var (
	testOnce      sync.Once
	testPersistor *keelmongo.Persistor
	testSetupErr  error
)

// testCollection returns a collection named after the calling test, backed by a
// mongo container shared across the package. The container is reaped by ryuk when
// the test binary exits.
func testCollection(t *testing.T) *keelmongo.Collection {
	t.Helper()

	testOnce.Do(func() {
		ctx := context.Background()

		container, err := mongodbcontainer.Run(ctx, testMongoImage)
		if err != nil {
			testSetupErr = err

			return
		}

		uri, err := container.ConnectionString(ctx)
		if err != nil {
			testSetupErr = err

			return
		}

		// ConnectionString already ends in "/", and keel reads the database name
		// from the uri path - concatenating naively yields a "//keeltest" path
		// that mongo rejects as an invalid namespace.
		testPersistor, testSetupErr = keelmongo.New(ctx, strings.TrimSuffix(uri, "/")+"/"+testDatabase)
	})

	require.NoError(t, testSetupErr, "failed to start mongo test container - is docker running?")

	col, err := testPersistor.Collection(strings.ReplaceAll(t.Name(), "/", "_"))
	require.NoError(t, err)

	return col
}

// --- tests -------------------------------------------------------------------

// TestUpsertBlindWritePreservesCreatedAt is the regression test for the bug this
// change fixes: a caller that never read the document cannot supply the stored
// creation timestamp, and must not destroy it by writing its own zero value.
func TestUpsertBlindWritePreservesCreatedAt(t *testing.T) {
	col := testCollection(t)
	ctx := t.Context()

	// first blind write inserts and stamps createdAt
	require.NoError(t, col.Upsert(ctx, "a", &insertOnlyEntity{ID: "a", Name: "first"}))

	var inserted insertOnlyEntity
	require.NoError(t, col.Get(ctx, "a", &inserted))
	require.NotZero(t, inserted.CreatedAt, "createdAt must be stamped on insert")
	assert.Equal(t, "first", inserted.Name)

	time.Sleep(2 * time.Millisecond)

	// second blind write updates - createdAt must survive
	require.NoError(t, col.Upsert(ctx, "a", &insertOnlyEntity{ID: "a", Name: "second"}))

	var updated insertOnlyEntity
	require.NoError(t, col.Get(ctx, "a", &updated))
	assert.Equal(t, inserted.CreatedAt, updated.CreatedAt, "createdAt must survive a blind update")
	assert.Equal(t, "second", updated.Name)
	assert.Greater(t, updated.UpdatedAt, inserted.UpdatedAt, "updatedAt must advance")
}

// TestUpsertKeepsCallerSuppliedCreatedAt covers the read-modify-write case: the
// entity already carries createdAt, so it travels in $set and InsertOnlyFields
// must stay silent - returning it as well would make mongo reject the write with
// a conflicting update path.
func TestUpsertKeepsCallerSuppliedCreatedAt(t *testing.T) {
	col := testCollection(t)
	ctx := t.Context()

	origin := newTimestamp(time.Date(2020, time.March, 1, 12, 0, 0, 0, time.UTC))

	require.NoError(t, col.Upsert(ctx, "b", &insertOnlyEntity{ID: "b", Name: "first", CreatedAt: origin}))

	var inserted insertOnlyEntity
	require.NoError(t, col.Get(ctx, "b", &inserted))
	assert.Equal(t, origin, inserted.CreatedAt, "a caller-supplied createdAt must be honoured")

	// updating with the same createdAt must not trip the $set/$setOnInsert conflict
	require.NoError(t, col.Upsert(ctx, "b", &insertOnlyEntity{ID: "b", Name: "second", CreatedAt: origin}))

	var updated insertOnlyEntity
	require.NoError(t, col.Get(ctx, "b", &updated))
	assert.Equal(t, origin, updated.CreatedAt)
	assert.Equal(t, "second", updated.Name)
}

// TestUpsertLegacyEntityBehaviourUnchanged pins the old behaviour for entities
// that do not implement EntityWithInsertOnlyFields, proving the change is not
// breaking for existing keel consumers. Such an entity still has its createdAt
// overwritten on a blind write - that is the pre-existing contract.
func TestUpsertLegacyEntityBehaviourUnchanged(t *testing.T) {
	col := testCollection(t)
	ctx := t.Context()

	require.NoError(t, col.Upsert(ctx, "c", &legacyEntity{ID: "c", Name: "first"}))

	var inserted legacyEntity
	require.NoError(t, col.Get(ctx, "c", &inserted))
	require.NotZero(t, inserted.CreatedAt)

	time.Sleep(2 * time.Millisecond)

	require.NoError(t, col.Upsert(ctx, "c", &legacyEntity{ID: "c", Name: "second"}))

	var updated legacyEntity
	require.NoError(t, col.Get(ctx, "c", &updated))
	assert.Greater(t, updated.CreatedAt, inserted.CreatedAt,
		"legacy entities keep the old behaviour: a blind write restamps createdAt")
}

// TestUpsertManyBlindWritePreservesCreatedAt is the bulk counterpart of
// TestUpsertBlindWritePreservesCreatedAt.
func TestUpsertManyBlindWritePreservesCreatedAt(t *testing.T) {
	col := testCollection(t)
	ctx := t.Context()

	require.NoError(t, col.UpsertMany(ctx, []keelmongo.Entity{
		&insertOnlyEntity{ID: "d1", Name: "first"},
		&insertOnlyEntity{ID: "d2", Name: "first"},
	}))

	var d1 insertOnlyEntity
	require.NoError(t, col.Get(ctx, "d1", &d1))
	require.NotZero(t, d1.CreatedAt)

	time.Sleep(2 * time.Millisecond)

	require.NoError(t, col.UpsertMany(ctx, []keelmongo.Entity{
		&insertOnlyEntity{ID: "d1", Name: "second"},
		&insertOnlyEntity{ID: "d2", Name: "second"},
	}))

	var updated insertOnlyEntity
	require.NoError(t, col.Get(ctx, "d1", &updated))
	assert.Equal(t, d1.CreatedAt, updated.CreatedAt, "createdAt must survive a blind bulk update")
	assert.Equal(t, "second", updated.Name)
	assert.Greater(t, updated.UpdatedAt, d1.UpdatedAt)
}

// --- document shape (no container required) ----------------------------------

// TestUpsertUpdateDocumentShape exercises the exported builder directly, so the
// policy is covered even where docker is unavailable.
func TestUpsertUpdateDocumentShape(t *testing.T) {
	now := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)

	t.Run("blind write defers createdAt to setOnInsert", func(t *testing.T) {
		update := keelmongo.UpsertUpdate(&insertOnlyEntity{ID: "a"}, now)

		require.Len(t, update, 2)
		assert.Equal(t, "$set", update[0].Key)
		assert.Equal(t, "$setOnInsert", update[1].Key)
		assert.Equal(t, bson.D{{Key: "createdAt", Value: newTimestamp(now)}}, update[1].Value)
	})

	t.Run("caller supplied createdAt stays in set", func(t *testing.T) {
		update := keelmongo.UpsertUpdate(&insertOnlyEntity{ID: "a", CreatedAt: 42}, now)

		require.Len(t, update, 1)
		assert.Equal(t, "$set", update[0].Key)
	})

	t.Run("legacy entity gets no setOnInsert", func(t *testing.T) {
		update := keelmongo.UpsertUpdate(&legacyEntity{ID: "a"}, now)

		require.Len(t, update, 1)
		assert.Equal(t, "$set", update[0].Key)
	})
}

// TestApplyTimestamps covers the companion helper's two branches.
func TestApplyTimestamps(t *testing.T) {
	now := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)

	t.Run("defers createdAt when the entity declares insert only fields", func(t *testing.T) {
		entity := &insertOnlyEntity{ID: "a"}

		keelmongo.ApplyTimestamps(entity, now)

		assert.Zero(t, entity.CreatedAt, "createdAt is left to $setOnInsert")
		assert.Equal(t, newTimestamp(now), entity.UpdatedAt)
	})

	t.Run("stamps createdAt for legacy entities", func(t *testing.T) {
		entity := &legacyEntity{ID: "a"}

		keelmongo.ApplyTimestamps(entity, now)

		assert.Equal(t, newTimestamp(now), entity.CreatedAt)
		assert.Equal(t, newTimestamp(now), entity.UpdatedAt)
	})
}
