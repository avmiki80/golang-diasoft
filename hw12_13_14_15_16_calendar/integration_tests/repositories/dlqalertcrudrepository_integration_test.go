//go:build integration
// +build integration

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	db2 "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDLQAlertCrudRepository_Create_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	alert := domain.DLQAlert{
		Topic:           "calendar-notification-command-dlq",
		OriginalTopic:   "calendar-notification-command",
		Partition:       0,
		MessageOffset:   12345,
		MessageKey:      "test-key-1",
		ErrorMessage:    "failed to process message",
		ErrorCount:      3,
		OriginalMessage: `{"payload":{"eventId":"123","title":"Test"}}`,
		FailedAt:        time.Now().UTC().Truncate(time.Second),
		Status:          "new",
	}

	createdAlert, err := repo.Create(ctx, db, alert)
	require.NoError(t, err)
	require.NotNil(t, createdAlert)
	assert.NotEmpty(t, createdAlert.ID, "ID should be auto-generated")
	assert.Equal(t, alert.Topic, createdAlert.Topic)
	assert.Equal(t, alert.OriginalTopic, createdAlert.OriginalTopic)
	assert.Equal(t, alert.Partition, createdAlert.Partition)
	assert.Equal(t, alert.MessageOffset, createdAlert.MessageOffset)
	assert.Equal(t, alert.ErrorMessage, createdAlert.ErrorMessage)
	assert.Equal(t, alert.ErrorCount, createdAlert.ErrorCount)
	assert.Equal(t, alert.Status, createdAlert.Status)
}

func TestDLQAlertCrudRepository_GetByID_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	alert := domain.DLQAlert{
		Topic:           "calendar-notification-command-dlq",
		OriginalTopic:   "calendar-notification-command",
		Partition:       0,
		MessageOffset:   12345,
		MessageKey:      "test-key-2",
		ErrorMessage:    "connection timeout",
		ErrorCount:      5,
		OriginalMessage: `{"payload":{"eventId":"456"}}`,
		FailedAt:        time.Now().UTC().Truncate(time.Second),
		Status:          "new",
	}

	createdAlert, err := repo.Create(ctx, db, alert)
	require.NoError(t, err)

	retrievedAlert, err := repo.GetByID(ctx, db, createdAlert.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedAlert)
	assert.Equal(t, createdAlert.ID, retrievedAlert.ID)
	assert.Equal(t, alert.Topic, retrievedAlert.Topic)
	assert.Equal(t, alert.OriginalTopic, retrievedAlert.OriginalTopic)
	assert.Equal(t, alert.ErrorMessage, retrievedAlert.ErrorMessage)
	assert.Equal(t, alert.ErrorCount, retrievedAlert.ErrorCount)
	assert.Equal(t, alert.Status, retrievedAlert.Status)
}

func TestDLQAlertCrudRepository_GetByID_NotFound_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	_, err := repo.GetByID(ctx, db, "550e8400-e29b-41d4-a716-446655440000")
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestDLQAlertCrudRepository_Update_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	alert := domain.DLQAlert{
		Topic:           "calendar-notification-command-dlq",
		OriginalTopic:   "calendar-notification-command",
		Partition:       0,
		MessageOffset:   12345,
		MessageKey:      "test-key-3",
		ErrorMessage:    "database error",
		ErrorCount:      2,
		OriginalMessage: `{"payload":{"eventId":"789"}}`,
		FailedAt:        time.Now().UTC().Truncate(time.Second),
		Status:          "new",
	}

	createdAlert, err := repo.Create(ctx, db, alert)
	require.NoError(t, err)

	// Update status to resolved
	createdAlert.Status = "resolved"
	updatedAlert, err := repo.Update(ctx, db, createdAlert.ID, *createdAlert)
	require.NoError(t, err)
	require.NotNil(t, updatedAlert)
	assert.Equal(t, "resolved", updatedAlert.Status)

	// Verify update persisted
	retrievedAlert, err := repo.GetByID(ctx, db, createdAlert.ID)
	require.NoError(t, err)
	assert.Equal(t, "resolved", retrievedAlert.Status)
}

func TestDLQAlertCrudRepository_Update_NotFound_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	alert := domain.DLQAlert{
		Status: "resolved",
	}

	_, err := repo.Update(ctx, db, "550e8400-e29b-41d4-a716-446655440000", alert)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestDLQAlertCrudRepository_Delete_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	alert := domain.DLQAlert{
		Topic:           "calendar-notification-command-dlq",
		OriginalTopic:   "calendar-notification-command",
		Partition:       0,
		MessageOffset:   12345,
		MessageKey:      "test-key-4",
		ErrorMessage:    "validation error",
		ErrorCount:      1,
		OriginalMessage: `{"payload":{}}`,
		FailedAt:        time.Now().UTC().Truncate(time.Second),
		Status:          "new",
	}

	createdAlert, err := repo.Create(ctx, db, alert)
	require.NoError(t, err)

	err = repo.Delete(ctx, db, createdAlert.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, db, createdAlert.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestDLQAlertCrudRepository_Delete_NotFound_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	err := repo.Delete(ctx, db, "550e8400-e29b-41d4-a716-446655440000")
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
}

func TestDLQAlertCrudRepository_GetUnresolvedCount_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	// Initially should be 0
	count, err := repo.GetUnresolvedCount(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create 3 new alerts
	for i := 0; i < 3; i++ {
		alert := domain.DLQAlert{
			Topic:           "calendar-notification-command-dlq",
			OriginalTopic:   "calendar-notification-command",
			Partition:       i,
			MessageOffset:   int64(i * 1000),
			MessageKey:      "test-key",
			ErrorMessage:    "error",
			ErrorCount:      1,
			OriginalMessage: `{}`,
			FailedAt:        time.Now().UTC(),
			Status:          "new",
		}
		_, err := repo.Create(ctx, db, alert)
		require.NoError(t, err)
	}

	// Should have 3 unresolved
	count, err = repo.GetUnresolvedCount(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Create 2 resolved alerts
	for i := 0; i < 2; i++ {
		alert := domain.DLQAlert{
			Topic:           "calendar-notification-command-dlq",
			OriginalTopic:   "calendar-notification-command",
			Partition:       i + 10,
			MessageOffset:   int64((i + 10) * 1000),
			MessageKey:      "test-key",
			ErrorMessage:    "error",
			ErrorCount:      1,
			OriginalMessage: `{}`,
			FailedAt:        time.Now().UTC(),
			Status:          "resolved",
		}
		_, err := repo.Create(ctx, db, alert)
		require.NoError(t, err)
	}

	// Should still have 3 unresolved (resolved don't count)
	count, err = repo.GetUnresolvedCount(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestDLQAlertCrudRepository_GetByStatus_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	// Create alerts with different statuses
	statuses := []string{"new", "new", "processing", "resolved", "new"}
	for i, status := range statuses {
		alert := domain.DLQAlert{
			Topic:           "calendar-notification-command-dlq",
			OriginalTopic:   "calendar-notification-command",
			Partition:       i,
			MessageOffset:   int64(i * 1000),
			MessageKey:      "test-key",
			ErrorMessage:    "error",
			ErrorCount:      1,
			OriginalMessage: `{}`,
			FailedAt:        time.Now().UTC().Add(time.Duration(i) * time.Second),
			Status:          status,
		}
		_, err := repo.Create(ctx, db, alert)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different created_at times
	}

	t.Run("get new alerts", func(t *testing.T) {
		alerts, err := repo.GetByStatus(ctx, db, "new", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 3)
		for _, alert := range alerts {
			assert.Equal(t, "new", alert.Status)
		}
	})

	t.Run("get processing alerts", func(t *testing.T) {
		alerts, err := repo.GetByStatus(ctx, db, "processing", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "processing", alerts[0].Status)
	})

	t.Run("get resolved alerts", func(t *testing.T) {
		alerts, err := repo.GetByStatus(ctx, db, "resolved", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "resolved", alerts[0].Status)
	})

	t.Run("limit results", func(t *testing.T) {
		alerts, err := repo.GetByStatus(ctx, db, "new", 2)
		require.NoError(t, err)
		assert.Len(t, alerts, 2)
	})

	t.Run("results ordered by created_at DESC", func(t *testing.T) {
		alerts, err := repo.GetByStatus(ctx, db, "new", 10)
		require.NoError(t, err)
		require.Len(t, alerts, 3)

		// Should be ordered DESC (newest first)
		for i := 1; i < len(alerts); i++ {
			assert.True(t, alerts[i-1].CreatedAt.After(alerts[i].CreatedAt) ||
				alerts[i-1].CreatedAt.Equal(alerts[i].CreatedAt),
				"Alerts should be ordered by created_at DESC")
		}
	})

	t.Run("no alerts with status", func(t *testing.T) {
		alerts, err := repo.GetByStatus(ctx, db, "failed", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 0)
	})
}

func TestDLQAlertCrudRepository_Transaction_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	repo := db2.NewDLQAlertCrudRepository(db)

	t.Run("commit transaction", func(t *testing.T) {
		tx, err := db.Beginx()
		require.NoError(t, err)

		alert := domain.DLQAlert{
			Topic:           "calendar-notification-command-dlq",
			OriginalTopic:   "calendar-notification-command",
			Partition:       0,
			MessageOffset:   99999,
			MessageKey:      "tx-test-key",
			ErrorMessage:    "tx error",
			ErrorCount:      1,
			OriginalMessage: `{}`,
			FailedAt:        time.Now().UTC(),
			Status:          "new",
		}

		createdAlert, err := repo.Create(ctx, tx, alert)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify alert exists after commit
		retrievedAlert, err := repo.GetByID(ctx, db, createdAlert.ID)
		require.NoError(t, err)
		assert.Equal(t, alert.MessageKey, retrievedAlert.MessageKey)
	})

	t.Run("rollback transaction", func(t *testing.T) {
		cleanupTestData(t, db)

		tx, err := db.Beginx()
		require.NoError(t, err)

		alert := domain.DLQAlert{
			Topic:           "calendar-notification-command-dlq",
			OriginalTopic:   "calendar-notification-command",
			Partition:       0,
			MessageOffset:   88888,
			MessageKey:      "rollback-test-key",
			ErrorMessage:    "rollback error",
			ErrorCount:      1,
			OriginalMessage: `{}`,
			FailedAt:        time.Now().UTC(),
			Status:          "new",
		}

		createdAlert, err := repo.Create(ctx, tx, alert)
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)

		// Verify alert doesn't exist after rollback
		_, err = repo.GetByID(ctx, db, createdAlert.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
	})
}
