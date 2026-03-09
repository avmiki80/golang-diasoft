//go:build integration
// +build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/testhelpers"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNotificationTestEnvironment(t *testing.T) (*NotificationTestEnvironment, func()) {
	t.Helper()

	pc := testhelpers.SetupPostgresContainer(t, "notification_service_test")
	txManager := database.NewTxManager(pc.DB)
	crudRepo := db.NewNotificationCrudRepository(pc.DB)
	repository, err := db.NewNotificationRepository(crudRepo)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	service := NewNotificationService(repository, txManager)

	env := &NotificationTestEnvironment{
		DB:        pc.DB,
		TxManager: txManager,
		Service:   service,
	}

	cleanup := func() {
		testhelpers.CleanupTestData(t, pc.DB)
	}

	return env, cleanup
}

type NotificationTestEnvironment struct {
	DB        *sqlx.DB
	TxManager database.TxManager
	Service   NotificationService
}

func TestNotificationService_CreateNotification_Success(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "Meeting Reminder",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	createdNotification, err := env.Service.CreateNotification(ctx, notification)

	require.NoError(t, err)
	require.NotNil(t, createdNotification)
	assert.NotEmpty(t, createdNotification.ID)
	assert.Equal(t, notification.Title, createdNotification.Title)
	assert.Equal(t, notification.UserID, createdNotification.UserID)
	assert.Equal(t, notification.EventID, createdNotification.EventID)
	assert.True(t, notification.StartDate.Equal(createdNotification.StartDate))
	assert.True(t, notification.EndDate.Equal(createdNotification.EndDate))
}

func TestNotificationService_CreateNotification_ValidationErrors(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New().String()
	eventID := uuid.New().String()

	tests := []struct {
		name         string
		notification domain.Notification
		expectedErr  error
	}{
		{
			name: "empty title",
			notification: domain.Notification{
				Title:     "",
				StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
				UserID:    userID,
				EventID:   eventID,
			},
			expectedErr: ErrInvalidEventTitle,
		},
		{
			name: "empty user ID",
			notification: domain.Notification{
				Title:     "Test Notification",
				StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
				UserID:    "",
				EventID:   eventID,
			},
			expectedErr: ErrInvalidUserID,
		},
		{
			name: "empty event ID",
			notification: domain.Notification{
				Title:     "Test Notification",
				StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
				UserID:    userID,
				EventID:   "",
			},
			expectedErr: ErrInvalidEventID,
		},
		{
			name: "zero start date",
			notification: domain.Notification{
				Title:     "Test Notification",
				StartDate: time.Time{},
				EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
				UserID:    userID,
				EventID:   eventID,
			},
			expectedErr: ErrInvalidStartDate,
		},
		{
			name: "zero end date",
			notification: domain.Notification{
				Title:     "Test Notification",
				StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				EndDate:   time.Time{},
				UserID:    userID,
				EventID:   eventID,
			},
			expectedErr: ErrInvalidEndDate,
		},
		{
			name: "end date before start date",
			notification: domain.Notification{
				Title:     "Test Notification",
				StartDate: time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:    userID,
				EventID:   eventID,
			},
			expectedErr: ErrInvalidDateRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := env.Service.CreateNotification(ctx, tt.notification)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestNotificationService_CreateNotification_Idempotent(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()
	eventID := uuid.New().String()

	notification := domain.Notification{
		Title:     "Idempotent Notification",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   eventID,
	}

	// Create first time
	created1, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)
	require.NotNil(t, created1)

	// Create second time with same event_id - should return existing
	created2, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)
	require.NotNil(t, created2)

	// Should return the same notification
	assert.Equal(t, created1.ID, created2.ID)
	assert.Equal(t, created1.EventID, created2.EventID)
}

func TestNotificationService_UpdateNotification_Success(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "Original Title",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	created, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)

	// Update notification
	updated := domain.Notification{
		Title:     "Updated Title",
		StartDate: time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 15, 0, 0, 0, time.UTC),
		UserID:    created.UserID,
		EventID:   created.EventID,
	}

	updatedNotification, err := env.Service.UpdateNotification(ctx, created.ID, updated)
	require.NoError(t, err)
	require.NotNil(t, updatedNotification)
	assert.Equal(t, "Updated Title", updatedNotification.Title)
	assert.True(t, updatedNotification.StartDate.Equal(time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC)))
}

func TestNotificationService_UpdateNotification_NotFound(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "Test",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	_, err := env.Service.UpdateNotification(ctx, uuid.New().String(), notification)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotFound)
}

func TestNotificationService_UpdateNotification_EmptyID(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "Test",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	_, err := env.Service.UpdateNotification(ctx, "", notification)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEventID)
}

func TestNotificationService_DeleteNotification_Success(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "To Be Deleted",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	created, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)

	err = env.Service.DeleteNotification(ctx, created.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = env.Service.GetNotificationByID(ctx, created.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotFound)
}

func TestNotificationService_DeleteNotification_NotFound(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	err := env.Service.DeleteNotification(ctx, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotFound)
}

func TestNotificationService_DeleteNotification_EmptyID(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	err := env.Service.DeleteNotification(ctx, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEventID)
}

func TestNotificationService_GetNotificationByID_Success(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "Test Notification",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	created, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)

	retrieved, err := env.Service.GetNotificationByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, notification.Title, retrieved.Title)
	assert.Equal(t, notification.UserID, retrieved.UserID)
	assert.Equal(t, notification.EventID, retrieved.EventID)
}

func TestNotificationService_GetNotificationByID_NotFound(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	_, err := env.Service.GetNotificationByID(ctx, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotFound)
}

func TestNotificationService_GetNotificationByID_EmptyID(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	_, err := env.Service.GetNotificationByID(ctx, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEventID)
}

func TestNotificationService_TransactionCommit(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	notification := domain.Notification{
		Title:     "Transaction Test",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	created, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)

	// Verify notification exists
	retrieved, err := env.Service.GetNotificationByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
}

func TestNotificationService_MultipleNotifications_SameEvent(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()
	eventID := uuid.New().String()

	notification1 := domain.Notification{
		Title:     "First Notification",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   eventID,
	}

	notification2 := domain.Notification{
		Title:     "Second Notification",
		StartDate: time.Date(2024, 3, 16, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 16, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   eventID,
	}

	// Create first notification
	created1, err := env.Service.CreateNotification(ctx, notification1)
	require.NoError(t, err)

	// Try to create second notification with same event_id - should return first one (idempotent)
	created2, err := env.Service.CreateNotification(ctx, notification2)
	require.NoError(t, err)

	// Due to idempotency, should return the same notification
	assert.Equal(t, created1.ID, created2.ID)
	assert.Equal(t, created1.EventID, created2.EventID)
}

func TestNotificationService_CRUD_Workflow(t *testing.T) {
	env, cleanup := setupNotificationTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// Create
	notification := domain.Notification{
		Title:     "Workflow Test",
		StartDate: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 11, 0, 0, 0, time.UTC),
		UserID:    uuid.New().String(),
		EventID:   uuid.New().String(),
	}

	created, err := env.Service.CreateNotification(ctx, notification)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)

	// Read
	retrieved, err := env.Service.GetNotificationByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "Workflow Test", retrieved.Title)

	// Update
	updated := domain.Notification{
		Title:     "Updated Workflow",
		StartDate: time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 15, 15, 0, 0, 0, time.UTC),
		UserID:    created.UserID,
		EventID:   created.EventID,
	}

	updatedNotification, err := env.Service.UpdateNotification(ctx, created.ID, updated)
	require.NoError(t, err)
	assert.Equal(t, "Updated Workflow", updatedNotification.Title)

	// Delete
	err = env.Service.DeleteNotification(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = env.Service.GetNotificationByID(ctx, created.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotFound)
}
