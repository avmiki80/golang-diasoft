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

func TestNotificationCrudRepository_Create_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	notification := domain.Notification{
		Title:     "Meeting Reminder",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		EventID:   "550e8400-e29b-41d4-a716-446655440010",
	}

	createdNotification, err := repo.Create(ctx, db, notification)
	require.NoError(t, err)
	require.NotNil(t, createdNotification)
	assert.NotEmpty(t, createdNotification.ID, "ID should be auto-generated")
	assert.Equal(t, notification.Title, createdNotification.Title)
	assert.Equal(t, notification.UserID, createdNotification.UserID)
	assert.Equal(t, notification.EventID, createdNotification.EventID)
	assert.True(t, notification.StartDate.Equal(createdNotification.StartDate))
	assert.True(t, notification.EndDate.Equal(createdNotification.EndDate))
}

func TestNotificationCrudRepository_GetByID_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	notification := domain.Notification{
		Title:     "Deadline Reminder",
		StartDate: time.Date(2024, 3, 20, 9, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440002",
		EventID:   "550e8400-e29b-41d4-a716-446655440011",
	}

	createdNotification, err := repo.Create(ctx, db, notification)
	require.NoError(t, err)

	retrievedNotification, err := repo.GetByID(ctx, db, createdNotification.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedNotification)
	assert.Equal(t, createdNotification.ID, retrievedNotification.ID)
	assert.Equal(t, notification.Title, retrievedNotification.Title)
	assert.Equal(t, notification.UserID, retrievedNotification.UserID)
	assert.Equal(t, notification.EventID, retrievedNotification.EventID)
}

func TestNotificationCrudRepository_GetByID_NotFound_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	_, err := repo.GetByID(ctx, db, "550e8400-e29b-41d4-a716-446655440000")
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestNotificationCrudRepository_Update_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	notification := domain.Notification{
		Title:     "Original Title",
		StartDate: time.Date(2024, 3, 25, 14, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440003",
		EventID:   "550e8400-e29b-41d4-a716-446655440012",
	}

	createdNotification, err := repo.Create(ctx, db, notification)
	require.NoError(t, err)

	// Update notification
	createdNotification.Title = "Updated Title"
	createdNotification.StartDate = time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC)
	createdNotification.EndDate = time.Date(2024, 3, 25, 16, 0, 0, 0, time.UTC)

	updatedNotification, err := repo.Update(ctx, db, createdNotification.ID, *createdNotification)
	require.NoError(t, err)
	require.NotNil(t, updatedNotification)
	assert.Equal(t, "Updated Title", updatedNotification.Title)

	// Verify update persisted
	retrievedNotification, err := repo.GetByID(ctx, db, createdNotification.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrievedNotification.Title)
	assert.True(t, retrievedNotification.StartDate.Equal(time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC)))
}

func TestNotificationCrudRepository_Update_NotFound_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	notification := domain.Notification{
		Title:     "Non-existent",
		StartDate: time.Now().UTC(),
		EndDate:   time.Now().UTC().Add(time.Hour),
		UserID:    "550e8400-e29b-41d4-a716-446655440023",
		EventID:   "550e8400-e29b-41d4-a716-446655440013",
	}

	_, err := repo.Update(ctx, db, "550e8400-e29b-41d4-a716-446655440000", notification)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestNotificationCrudRepository_Delete_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	notification := domain.Notification{
		Title:     "To Be Deleted",
		StartDate: time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 4, 1, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440004",
		EventID:   "550e8400-e29b-41d4-a716-446655440014",
	}

	createdNotification, err := repo.Create(ctx, db, notification)
	require.NoError(t, err)

	err = repo.Delete(ctx, db, createdNotification.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, db, createdNotification.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestNotificationCrudRepository_Delete_NotFound_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	err := repo.Delete(ctx, db, "550e8400-e29b-41d4-a716-446655440000")
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestNotificationCrudRepository_Transaction_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	t.Run("commit transaction", func(t *testing.T) {
		tx, err := db.Beginx()
		require.NoError(t, err)

		notification := domain.Notification{
			Title:     "Transaction Test",
			StartDate: time.Date(2024, 4, 5, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 4, 5, 11, 0, 0, 0, time.UTC),
			UserID:    "550e8400-e29b-41d4-a716-446655440020",
			EventID:   "550e8400-e29b-41d4-a716-446655440015",
		}

		createdNotification, err := repo.Create(ctx, tx, notification)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify notification exists after commit
		retrievedNotification, err := repo.GetByID(ctx, db, createdNotification.ID)
		require.NoError(t, err)
		assert.Equal(t, notification.Title, retrievedNotification.Title)
	})

	t.Run("rollback transaction", func(t *testing.T) {
		cleanupTestData(t, db)

		tx, err := db.Beginx()
		require.NoError(t, err)

		notification := domain.Notification{
			Title:     "Rollback Test",
			StartDate: time.Date(2024, 4, 6, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 4, 6, 11, 0, 0, 0, time.UTC),
			UserID:    "550e8400-e29b-41d4-a716-446655440021",
			EventID:   "550e8400-e29b-41d4-a716-446655440016",
		}

		createdNotification, err := repo.Create(ctx, tx, notification)
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)

		// Verify notification doesn't exist after rollback
		_, err = repo.GetByID(ctx, db, createdNotification.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
	})
}

func TestNotificationCrudRepository_MultipleNotifications_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := NewNotificationCrudRepository(db)

	// Create multiple notifications for the same event
	eventID := "550e8400-e29b-41d4-a716-446655440017"
	userID := "550e8400-e29b-41d4-a716-446655440022"

	notification1 := domain.Notification{
		Title:     "First Notification",
		StartDate: time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 5, 1, 11, 0, 0, 0, time.UTC),
		UserID:    userID,
		EventID:   eventID,
	}

	notification2 := domain.Notification{
		Title:     "Second Notification",
		StartDate: time.Date(2024, 5, 2, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 5, 2, 11, 0, 0, 0, time.UTC),
		UserID:    userID,
		EventID:   eventID,
	}

	created1, err := repo.Create(ctx, db, notification1)
	require.NoError(t, err)

	created2, err := repo.Create(ctx, db, notification2)
	require.NoError(t, err)

	// Verify both exist and have different IDs
	assert.NotEqual(t, created1.ID, created2.ID)

	retrieved1, err := repo.GetByID(ctx, db, created1.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Notification", retrieved1.Title)

	retrieved2, err := repo.GetByID(ctx, db, created2.ID)
	require.NoError(t, err)
	assert.Equal(t, "Second Notification", retrieved2.Title)
}
