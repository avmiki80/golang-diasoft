//go:build integration
// +build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/integration_tests/testhelpers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	services2 "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDLQAlertTestEnvironment(t *testing.T) (*DLQAlertTestEnvironment, func()) {
	t.Helper()

	pc := testhelpers.SetupPostgresContainer(t, "dlq_alert_service_test")
	txManager := database.NewTxManager(pc.DB)
	repository := db.NewDLQAlertCrudRepository(pc.DB)
	service := services2.NewDLQAlertService(repository, txManager)

	env := &DLQAlertTestEnvironment{
		DB:        pc.DB,
		TxManager: txManager,
		Service:   service,
	}

	cleanup := func() {
		testhelpers.CleanupTestData(t, pc.DB)
	}

	return env, cleanup
}

type DLQAlertTestEnvironment struct {
	DB        *sqlx.DB
	TxManager database.TxManager
	Service   services2.DLQAlertService
}

func TestDLQAlertService_CreateAlert_Success(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	alert := domain.DLQAlert{
		Topic:           "calendar-notification-command-dlq",
		OriginalTopic:   "calendar-notification-command",
		Partition:       0,
		MessageOffset:   12345,
		MessageKey:      "test-key-1",
		ErrorMessage:    "failed to process message",
		ErrorCount:      3,
		OriginalMessage: `{"payload":{"eventId":"550e8400-e29b-41d4-a716-446655440001"}}`,
		FailedAt:        time.Now().UTC().Truncate(time.Second),
		Status:          "new",
	}

	createdAlert, err := env.Service.CreateAlert(ctx, alert)

	require.NoError(t, err)
	require.NotNil(t, createdAlert)
	assert.NotEmpty(t, createdAlert.ID)
	assert.Equal(t, alert.Topic, createdAlert.Topic)
	assert.Equal(t, alert.OriginalTopic, createdAlert.OriginalTopic)
	assert.Equal(t, alert.ErrorMessage, createdAlert.ErrorMessage)
	assert.Equal(t, alert.ErrorCount, createdAlert.ErrorCount)
	assert.Equal(t, alert.Status, createdAlert.Status)
}

func TestDLQAlertService_GetUnresolvedCount_Empty(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	count, err := env.Service.GetUnresolvedCount(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDLQAlertService_GetUnresolvedCount_WithAlerts(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

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
		_, err := env.Service.CreateAlert(ctx, alert)
		require.NoError(t, err)
	}

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
		_, err := env.Service.CreateAlert(ctx, alert)
		require.NoError(t, err)
	}

	count, err := env.Service.GetUnresolvedCount(ctx)

	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestDLQAlertService_GetByStatus_Success(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

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
		_, err := env.Service.CreateAlert(ctx, alert)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	t.Run("get new alerts", func(t *testing.T) {
		alerts, err := env.Service.GetByStatus(ctx, "new", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 3)
		for _, alert := range alerts {
			assert.Equal(t, "new", alert.Status)
		}
	})

	t.Run("get processing alerts", func(t *testing.T) {
		alerts, err := env.Service.GetByStatus(ctx, "processing", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "processing", alerts[0].Status)
	})

	t.Run("get resolved alerts", func(t *testing.T) {
		alerts, err := env.Service.GetByStatus(ctx, "resolved", 10)
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "resolved", alerts[0].Status)
	})

	t.Run("limit results", func(t *testing.T) {
		alerts, err := env.Service.GetByStatus(ctx, "new", 2)
		require.NoError(t, err)
		assert.Len(t, alerts, 2)
	})
}

func TestDLQAlertService_UpdateAlertStatus_Success(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	alert := domain.DLQAlert{
		Topic:           "calendar-notification-command-dlq",
		OriginalTopic:   "calendar-notification-command",
		Partition:       0,
		MessageOffset:   12345,
		MessageKey:      "test-key",
		ErrorMessage:    "error",
		ErrorCount:      1,
		OriginalMessage: `{}`,
		FailedAt:        time.Now().UTC(),
		Status:          "new",
	}

	createdAlert, err := env.Service.CreateAlert(ctx, alert)
	require.NoError(t, err)

	// Update status to processing
	err = env.Service.UpdateAlertStatus(ctx, createdAlert.ID, "processing")
	require.NoError(t, err)

	// Verify status updated
	alerts, err := env.Service.GetByStatus(ctx, "processing", 10)
	require.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, createdAlert.ID, alerts[0].ID)
	assert.Equal(t, "processing", alerts[0].Status)

	// Update status to resolved
	err = env.Service.UpdateAlertStatus(ctx, createdAlert.ID, "resolved")
	require.NoError(t, err)

	// Verify status updated
	alerts, err = env.Service.GetByStatus(ctx, "resolved", 10)
	require.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, createdAlert.ID, alerts[0].ID)
	assert.Equal(t, "resolved", alerts[0].Status)
}

func TestDLQAlertService_UpdateAlertStatus_NotFound(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	err := env.Service.UpdateAlertStatus(ctx, "550e8400-e29b-41d4-a716-446655440000", "resolved")
	require.Error(t, err)
}

func TestDLQAlertService_TransactionCommit(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

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

	createdAlert, err := env.Service.CreateAlert(ctx, alert)
	require.NoError(t, err)
	assert.NotEmpty(t, createdAlert.ID)

	// Verify alert exists
	count, err := env.Service.GetUnresolvedCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDLQAlertService_MultipleAlerts_OrderedByCreatedAt(t *testing.T) {
	env, cleanup := setupDLQAlertTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// Create alerts with delays to ensure different created_at
	for i := 0; i < 5; i++ {
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
		_, err := env.Service.CreateAlert(ctx, alert)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	alerts, err := env.Service.GetByStatus(ctx, "new", 10)
	require.NoError(t, err)
	assert.Len(t, alerts, 5)

	// Verify ordered by created_at DESC (newest first)
	for i := 1; i < len(alerts); i++ {
		assert.True(t, alerts[i-1].CreatedAt.After(alerts[i].CreatedAt) ||
			alerts[i-1].CreatedAt.Equal(alerts[i].CreatedAt),
			"Alerts should be ordered by created_at DESC")
	}
}
