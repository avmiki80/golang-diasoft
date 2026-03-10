//go:build integration
// +build integration

package db

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository_FindNotification_ByEventID_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	// Create notifications for different events
	notification1 := domain.Notification{
		Title:     "Event 1 Notification",
		StartDate: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440030",
		EventID:   "550e8400-e29b-41d4-a716-446655440040",
	}

	notification2 := domain.Notification{
		Title:     "Event 2 Notification",
		StartDate: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 2, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440030",
		EventID:   "550e8400-e29b-41d4-a716-446655440041",
	}

	notification3 := domain.Notification{
		Title:     "Event 1 Another Notification",
		StartDate: time.Date(2024, 6, 3, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 3, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440031",
		EventID:   "550e8400-e29b-41d4-a716-446655440040",
	}

	_, err = repo.Create(ctx, db, notification1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, notification2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, notification3)
	require.NoError(t, err)

	t.Run("find by event_id", func(t *testing.T) {
		eventID := "550e8400-e29b-41d4-a716-446655440040"
		notifications, err := repo.FindNotification(ctx, db, &eventID, nil)
		require.NoError(t, err)
		assert.Len(t, notifications, 2)

		titles := make(map[string]bool)
		for _, n := range notifications {
			titles[n.Title] = true
			assert.Equal(t, eventID, n.EventID)
		}
		assert.True(t, titles["Event 1 Notification"])
		assert.True(t, titles["Event 1 Another Notification"])
	})

	t.Run("find by different event_id", func(t *testing.T) {
		eventID := "550e8400-e29b-41d4-a716-446655440041"
		notifications, err := repo.FindNotification(ctx, db, &eventID, nil)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, "Event 2 Notification", notifications[0].Title)
		assert.Equal(t, eventID, notifications[0].EventID)
	})

	t.Run("find by non-existent event_id", func(t *testing.T) {
		eventID := "550e8400-e29b-41d4-a716-446655440099"
		notifications, err := repo.FindNotification(ctx, db, &eventID, nil)
		require.NoError(t, err)
		assert.Len(t, notifications, 0)
	})
}

func TestNotificationRepository_FindNotification_ByUserID_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	// Create notifications for different users
	notification1 := domain.Notification{
		Title:     "User 1 Notification 1",
		StartDate: time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 7, 1, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440032",
		EventID:   "550e8400-e29b-41d4-a716-446655440042",
	}

	notification2 := domain.Notification{
		Title:     "User 1 Notification 2",
		StartDate: time.Date(2024, 7, 2, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 7, 2, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440032",
		EventID:   "550e8400-e29b-41d4-a716-446655440043",
	}

	notification3 := domain.Notification{
		Title:     "User 2 Notification",
		StartDate: time.Date(2024, 7, 3, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 7, 3, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440033",
		EventID:   "550e8400-e29b-41d4-a716-446655440044",
	}

	_, err = repo.Create(ctx, db, notification1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, notification2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, notification3)
	require.NoError(t, err)

	t.Run("find by user_id", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440032"
		notifications, err := repo.FindNotification(ctx, db, nil, &userID)
		require.NoError(t, err)
		assert.Len(t, notifications, 2)

		for _, n := range notifications {
			assert.Equal(t, userID, n.UserID)
		}
	})

	t.Run("find by different user_id", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440033"
		notifications, err := repo.FindNotification(ctx, db, nil, &userID)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, "User 2 Notification", notifications[0].Title)
	})

	t.Run("find by non-existent user_id", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440098"
		notifications, err := repo.FindNotification(ctx, db, nil, &userID)
		require.NoError(t, err)
		assert.Len(t, notifications, 0)
	})
}

func TestNotificationRepository_FindNotification_ByEventIDAndUserID_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	// Create notifications with various combinations
	notification1 := domain.Notification{
		Title:     "User 1 Event A",
		StartDate: time.Date(2024, 8, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 8, 1, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440034",
		EventID:   "550e8400-e29b-41d4-a716-446655440045",
	}

	notification2 := domain.Notification{
		Title:     "User 1 Event B",
		StartDate: time.Date(2024, 8, 2, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 8, 2, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440034",
		EventID:   "550e8400-e29b-41d4-a716-446655440046",
	}

	notification3 := domain.Notification{
		Title:     "User 2 Event A",
		StartDate: time.Date(2024, 8, 3, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 8, 3, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440035",
		EventID:   "550e8400-e29b-41d4-a716-446655440045",
	}

	_, err = repo.Create(ctx, db, notification1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, notification2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, notification3)
	require.NoError(t, err)

	t.Run("find by both event_id and user_id", func(t *testing.T) {
		eventID := "550e8400-e29b-41d4-a716-446655440045"
		userID := "550e8400-e29b-41d4-a716-446655440034"
		notifications, err := repo.FindNotification(ctx, db, &eventID, &userID)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, "User 1 Event A", notifications[0].Title)
		assert.Equal(t, eventID, notifications[0].EventID)
		assert.Equal(t, userID, notifications[0].UserID)
	})

	t.Run("find by event_id and different user_id", func(t *testing.T) {
		eventID := "550e8400-e29b-41d4-a716-446655440045"
		userID := "550e8400-e29b-41d4-a716-446655440035"
		notifications, err := repo.FindNotification(ctx, db, &eventID, &userID)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, "User 2 Event A", notifications[0].Title)
	})

	t.Run("no match for combination", func(t *testing.T) {
		eventID := "550e8400-e29b-41d4-a716-446655440046"
		userID := "550e8400-e29b-41d4-a716-446655440035"
		notifications, err := repo.FindNotification(ctx, db, &eventID, &userID)
		require.NoError(t, err)
		assert.Len(t, notifications, 0)
	})
}

func TestNotificationRepository_FindNotification_NoFilters_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	// Create several notifications
	for i := 0; i < 5; i++ {
		notification := domain.Notification{
			Title:     "Notification",
			StartDate: time.Date(2024, 9, i+1, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 9, i+1, 11, 0, 0, 0, time.UTC),
			UserID:    "550e8400-e29b-41d4-a716-446655440036",
			EventID:   "550e8400-e29b-41d4-a716-446655440047",
		}
		_, err := repo.Create(ctx, db, notification)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different created_at
	}

	t.Run("find all notifications without filters", func(t *testing.T) {
		notifications, err := repo.FindNotification(ctx, db, nil, nil)
		require.NoError(t, err)
		assert.Len(t, notifications, 5)
	})

	t.Run("results ordered by created_at", func(t *testing.T) {
		notifications, err := repo.FindNotification(ctx, db, nil, nil)
		require.NoError(t, err)
		require.Len(t, notifications, 5)

		// Should be ordered by created_at ASC
		for i := 1; i < len(notifications); i++ {
			assert.True(t, notifications[i-1].CreatedAt.Before(notifications[i].CreatedAt) ||
				notifications[i-1].CreatedAt.Equal(notifications[i].CreatedAt),
				"Notifications should be ordered by created_at ASC")
		}
	})
}

func TestNotificationRepository_FindNotification_EmptyStrings_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	notification := domain.Notification{
		Title:     "Test Notification",
		StartDate: time.Date(2024, 10, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 10, 1, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440037",
		EventID:   "550e8400-e29b-41d4-a716-446655440048",
	}

	_, err = repo.Create(ctx, db, notification)
	require.NoError(t, err)

	t.Run("empty event_id string treated as nil", func(t *testing.T) {
		emptyEventID := ""
		notifications, err := repo.FindNotification(ctx, db, &emptyEventID, nil)
		require.NoError(t, err)
		// Should return all notifications (empty string ignored)
		assert.GreaterOrEqual(t, len(notifications), 1)
	})

	t.Run("empty user_id string treated as nil", func(t *testing.T) {
		emptyUserID := ""
		notifications, err := repo.FindNotification(ctx, db, nil, &emptyUserID)
		require.NoError(t, err)
		// Should return all notifications (empty string ignored)
		assert.GreaterOrEqual(t, len(notifications), 1)
	})
}

func TestNotificationRepository_CRUD_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	notification := domain.Notification{
		Title:     "CRUD Test Notification",
		StartDate: time.Date(2024, 11, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 11, 1, 11, 0, 0, 0, time.UTC),
		UserID:    "550e8400-e29b-41d4-a716-446655440038",
		EventID:   "550e8400-e29b-41d4-a716-446655440049",
	}

	var createdNotificationID string

	t.Run("create and retrieve", func(t *testing.T) {
		createdNotification, err := repo.Create(ctx, db, notification)
		require.NoError(t, err)
		require.NotNil(t, createdNotification)
		assert.NotEmpty(t, createdNotification.ID)

		createdNotificationID = createdNotification.ID

		retrieved, err := repo.GetByID(ctx, db, createdNotification.ID)
		require.NoError(t, err)
		assert.Equal(t, createdNotification.ID, retrieved.ID)
		assert.Equal(t, notification.Title, retrieved.Title)
	})

	t.Run("update notification", func(t *testing.T) {
		notification.Title = "Updated CRUD Title"
		updatedNotification, err := repo.Update(ctx, db, createdNotificationID, notification)
		require.NoError(t, err)
		require.NotNil(t, updatedNotification)

		retrieved, err := repo.GetByID(ctx, db, createdNotificationID)
		require.NoError(t, err)
		assert.Equal(t, "Updated CRUD Title", retrieved.Title)
	})

	t.Run("delete notification", func(t *testing.T) {
		err := repo.Delete(ctx, db, createdNotificationID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, db, createdNotificationID)
		require.Error(t, err)
	})
}

func TestNotificationRepository_Idempotency_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewNotificationCrudRepository(db)
	repo, err := NewNotificationRepository(crudRepo)
	require.NoError(t, err)

	eventID := "550e8400-e29b-41d4-a716-446655440050"
	userID := "550e8400-e29b-41d4-a716-446655440039"

	notification := domain.Notification{
		Title:     "Idempotency Test",
		StartDate: time.Date(2024, 12, 1, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 12, 1, 11, 0, 0, 0, time.UTC),
		UserID:    userID,
		EventID:   eventID,
	}

	t.Run("create notification", func(t *testing.T) {
		_, err := repo.Create(ctx, db, notification)
		require.NoError(t, err)
	})

	t.Run("check notification exists by event_id", func(t *testing.T) {
		notifications, err := repo.FindNotification(ctx, db, &eventID, nil)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, eventID, notifications[0].EventID)
	})

	t.Run("attempt to create duplicate should be detectable", func(t *testing.T) {
		// Before creating, check if exists
		notifications, err := repo.FindNotification(ctx, db, &eventID, nil)
		require.NoError(t, err)

		if len(notifications) > 0 {
			// Already exists, use existing
			assert.Equal(t, "Idempotency Test", notifications[0].Title)
		} else {
			// Doesn't exist, create new
			_, err := repo.Create(ctx, db, notification)
			require.NoError(t, err)
		}

		// Verify still only one notification
		notifications, err = repo.FindNotification(ctx, db, &eventID, nil)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
	})
}
