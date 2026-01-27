//go:build integration
// +build integration

package db

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventCrudRepository_Create_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewEventCrudRepository(db)

	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(time.Hour).Truncate(time.Second),
		Description: "Test Description",
		UserID:      "550e8400-e29b-41d4-a716-446655440001",
		OffsetTime:  0,
	}

	t.Run("successful create", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)
		require.NotNil(t, createdEvent)
		assert.NotEmpty(t, createdEvent.ID, "ID should be auto-generated")

		// Verify event was created
		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.Equal(t, createdEvent.ID, retrieved.ID)
		assert.Equal(t, event.Title, retrieved.Title)
		assert.Equal(t, event.UserID, retrieved.UserID)
	})
}

func TestEventCrudRepository_GetByID_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewEventCrudRepository(db)

	event := domain.Event{
		Title:       "Test Event GetByID",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(time.Hour).Truncate(time.Second),
		Description: "Test Description",
		UserID:      "550e8400-e29b-41d4-a716-446655440002",
		OffsetTime:  0,
	}

	t.Run("get existing event", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, createdEvent.ID, retrieved.ID)
		assert.Equal(t, event.Title, retrieved.Title)
	})

	t.Run("get non-existing event", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, db, "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
		assert.Nil(t, retrieved)
	})
}

func TestEventCrudRepository_Update_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewEventCrudRepository(db)

	event := domain.Event{
		Title:       "Original Title",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(time.Hour).Truncate(time.Second),
		Description: "Original Description",
		UserID:      "550e8400-e29b-41d4-a716-446655440003",
		OffsetTime:  0,
	}

	t.Run("update existing event", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		updatedEvent := event
		updatedEvent.Title = "Updated Title"
		updatedEvent.Description = "Updated Description"

		result, err := repo.Update(ctx, db, createdEvent.ID, updatedEvent)
		require.NoError(t, err)
		require.NotNil(t, result)

		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)
		assert.Equal(t, "Updated Description", retrieved.Description)
	})

	t.Run("update non-existing event", func(t *testing.T) {
		result, err := repo.Update(ctx, db, "00000000-0000-0000-0000-000000000000", event)
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
		assert.Nil(t, result)
	})
}

func TestEventCrudRepository_Delete_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewEventCrudRepository(db)

	event := domain.Event{
		Title:       "Event to Delete",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(time.Hour).Truncate(time.Second),
		Description: "Will be deleted",
		UserID:      "550e8400-e29b-41d4-a716-446655440004",
		OffsetTime:  0,
	}

	t.Run("delete existing event", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		err = repo.Delete(ctx, db, createdEvent.ID)
		require.NoError(t, err)

		// Verify event was deleted
		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existing event", func(t *testing.T) {
		err := repo.Delete(ctx, db, "00000000-0000-0000-0000-000000000000")
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
	})
}

func TestEventCrudRepository_Transaction_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewEventCrudRepository(db)

	event1 := domain.Event{
		Title:       "Transaction Event 1",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(time.Hour).Truncate(time.Second),
		Description: "First event in transaction",
		UserID:      "550e8400-e29b-41d4-a716-446655440099",
		OffsetTime:  0,
	}

	event2 := domain.Event{
		Title:       "Transaction Event 2",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(2 * time.Hour).Truncate(time.Second),
		Description: "Second event in transaction",
		UserID:      "550e8400-e29b-41d4-a716-446655440099",
		OffsetTime:  0,
	}

	t.Run("commit transaction", func(t *testing.T) {
		tx, err := db.Beginx()
		require.NoError(t, err)

		_, err = repo.Create(ctx, tx, event1)
		require.NoError(t, err)

		_, err = repo.Create(ctx, tx, event2)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM public.events WHERE user_id = $1", "550e8400-e29b-41d4-a716-446655440099")
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("rollback transaction", func(t *testing.T) {
		cleanupTestData(t, db)

		tx, err := db.Beginx()
		require.NoError(t, err)

		_, err = repo.Create(ctx, tx, event1)
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)

		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM public.events WHERE user_id = $1", "550e8400-e29b-41d4-a716-446655440099")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
