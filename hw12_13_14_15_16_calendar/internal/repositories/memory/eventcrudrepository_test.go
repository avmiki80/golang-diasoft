package memory

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventCrudRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewEventCrudRepository()

	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Hour),
		Description: "Test Description",
		UserID:      "user-1",
		OffsetTime:  0,
	}

	t.Run("successful create", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)
		require.NotNil(t, createdEvent)
		assert.NotEmpty(t, createdEvent.ID, "ID should be auto-generated")

		// Verify event was created
		retrieved, err := repo.GetByID(ctx, nil, createdEvent.ID)
		require.NoError(t, err)
		assert.Equal(t, createdEvent.ID, retrieved.ID)
		assert.Equal(t, event.Title, retrieved.Title)
		assert.Equal(t, event.UserID, retrieved.UserID)
	})

	t.Run("multiple creates generate different IDs", func(t *testing.T) {
		event1, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)
		event2, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)
		assert.NotEqual(t, event1.ID, event2.ID, "Each create should generate unique ID")
	})
}

func TestEventCrudRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewEventCrudRepository()

	event := domain.Event{
		Title:       "Test Event 2",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Hour),
		Description: "Test Description 2",
		UserID:      "user-2",
		OffsetTime:  0,
	}

	t.Run("get existing event", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, nil, createdEvent.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, createdEvent.ID, retrieved.ID)
		assert.Equal(t, event.Title, retrieved.Title)
	})

	t.Run("get non-existing event", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, nil, "non-existing-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
		assert.Nil(t, retrieved)
	})
}

func TestEventCrudRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewEventCrudRepository()

	event := domain.Event{
		Title:       "Original Title",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Hour),
		Description: "Original Description",
		UserID:      "user-3",
		OffsetTime:  0,
	}

	t.Run("update existing event", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)

		updatedEvent := event
		updatedEvent.Title = "Updated Title"
		updatedEvent.Description = "Updated Description"

		result, err := repo.Update(ctx, nil, createdEvent.ID, updatedEvent)
		require.NoError(t, err)
		require.NotNil(t, result)

		retrieved, err := repo.GetByID(ctx, nil, createdEvent.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)
		assert.Equal(t, "Updated Description", retrieved.Description)
	})

	t.Run("update non-existing event", func(t *testing.T) {
		result, err := repo.Update(ctx, nil, "non-existing-id", event)
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
		assert.Nil(t, result)
	})
}

func TestEventCrudRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewEventCrudRepository()

	event := domain.Event{
		Title:       "Event to Delete",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Hour),
		Description: "Will be deleted",
		UserID:      "user-4",
		OffsetTime:  0,
	}

	t.Run("delete existing event", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)

		err = repo.Delete(ctx, nil, createdEvent.ID)
		require.NoError(t, err)

		// Verify event was deleted
		retrieved, err := repo.GetByID(ctx, nil, createdEvent.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existing event", func(t *testing.T) {
		err := repo.Delete(ctx, nil, "non-existing-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
	})
}

func TestEventCrudRepository_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewEventCrudRepository()

	// Test concurrent writes
	t.Run("concurrent creates", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				event := domain.Event{
					Title:       "Concurrent Event",
					StartDate:   time.Now(),
					EndDate:     time.Now().Add(time.Hour),
					Description: "Concurrent test",
					UserID:      "user-concurrent",
					OffsetTime:  0,
				}
				_, _ = repo.Create(ctx, nil, event)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all 10 events were created with unique IDs
		assert.Equal(t, 10, len(repo.events))
	})
}
