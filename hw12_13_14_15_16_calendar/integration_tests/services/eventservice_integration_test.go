//go:build integration
// +build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	services2 "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventService_CreateEvent_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test Description",
		UserID:      uuid.New().String(),
		OffsetTime:  0,
	}

	createdEvent, err := env.Service.CreateEvent(ctx, event)

	require.NoError(t, err)
	require.NotNil(t, createdEvent)
	assert.NotEmpty(t, createdEvent.ID)
	assert.Equal(t, event.Title, createdEvent.Title)
	assert.Equal(t, event.StartDate, createdEvent.StartDate)
	assert.Equal(t, event.EndDate, createdEvent.EndDate)
	assert.Equal(t, event.Description, createdEvent.Description)
	assert.Equal(t, event.UserID, createdEvent.UserID)
}

func TestEventService_CreateEvent_ValidationErrors(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	tests := []struct {
		name        string
		event       domain.Event
		expectedErr error
	}{
		{
			name: "empty title",
			event: domain.Event{
				Title:      "",
				StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
				UserID:     userID,
				OffsetTime: 0,
			},
			expectedErr: services2.ErrInvalidEventTitle,
		},
		{
			name: "empty user ID",
			event: domain.Event{
				Title:      "Test Event",
				StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
				UserID:     "",
				OffsetTime: 0,
			},
			expectedErr: services2.ErrInvalidUserID,
		},
		{
			name: "zero start date",
			event: domain.Event{
				Title:      "Test Event",
				StartDate:  time.Time{},
				EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
				UserID:     userID,
				OffsetTime: 0,
			},
			expectedErr: services2.ErrInvalidStartDate,
		},
		{
			name: "zero end date",
			event: domain.Event{
				Title:      "Test Event",
				StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndDate:    time.Time{},
				UserID:     userID,
				OffsetTime: 0,
			},
			expectedErr: services2.ErrInvalidEndDate,
		},
		{
			name: "end date before start date",
			event: domain.Event{
				Title:      "Test Event",
				StartDate:  time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				UserID:     userID,
				OffsetTime: 0,
			},
			expectedErr: services2.ErrInvalidDateRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := env.Service.CreateEvent(ctx, tt.event)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestEventService_CreateEvent_DateBusy(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Создаем первое событие
	event1 := domain.Event{
		Title:       "Event 1",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "First event",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, event1)
	require.NoError(t, err)

	// Пытаемся создать пересекающееся событие
	event2 := domain.Event{
		Title:       "Event 2",
		StartDate:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC),
		Description: "Overlapping event",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err = env.Service.CreateEvent(ctx, event2)
	assert.ErrorIs(t, err, services2.ErrDateBusy)
}

func TestEventService_UpdateEvent_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Создаем событие
	event := domain.Event{
		Title:       "Original Title",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Original Description",
		UserID:      userID,
		OffsetTime:  0,
	}

	created, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Обновляем событие
	updatedEvent := domain.Event{
		Title:       "Updated Title",
		StartDate:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		Description: "Updated Description",
		UserID:      userID,
		OffsetTime:  30,
	}

	result, err := env.Service.UpdateEvent(ctx, created.ID, updatedEvent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "Updated Title", result.Title)
	assert.Equal(t, "Updated Description", result.Description)
	assert.Equal(t, time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), result.StartDate)
	assert.Equal(t, time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC), result.EndDate)
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	event := domain.Event{
		Title:      "Test Event",
		StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		UserID:     uuid.New().String(),
		OffsetTime: 0,
	}

	_, err := env.Service.UpdateEvent(ctx, uuid.New().String(), event)
	assert.ErrorIs(t, err, services2.ErrEventNotFound)
}

func TestEventService_UpdateEvent_EmptyID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	event := domain.Event{
		Title:      "Test Event",
		StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		UserID:     uuid.New().String(),
		OffsetTime: 0,
	}

	_, err := env.Service.UpdateEvent(ctx, "", event)
	assert.ErrorIs(t, err, services2.ErrInvalidEventID)
}

func TestEventService_DeleteEvent_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	// Создаем событие
	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test Description",
		UserID:      uuid.New().String(),
		OffsetTime:  0,
	}

	created, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Удаляем событие
	err = env.Service.DeleteEvent(ctx, created.ID)
	require.NoError(t, err)

	// Проверяем, что событие удалено
	result, err := env.Service.GetEventByID(ctx, created.ID)
	assert.ErrorIs(t, err, services2.ErrEventNotFound)
	assert.Nil(t, result)
}

func TestEventService_DeleteEvent_EmptyID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	err := env.Service.DeleteEvent(ctx, "")
	assert.ErrorIs(t, err, services2.ErrInvalidEventID)
}

func TestEventService_GetEventByID_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	// Создаем событие
	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test Description",
		UserID:      uuid.New().String(),
		OffsetTime:  0,
	}

	created, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Получаем событие по ID
	result, err := env.Service.GetEventByID(ctx, created.ID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, created.Title, result.Title)
	assert.Equal(t, created.Description, result.Description)
}

func TestEventService_GetEventByID_NotFound(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	result, err := env.Service.GetEventByID(ctx, uuid.New().String())

	assert.ErrorIs(t, err, services2.ErrEventNotFound)
	assert.Nil(t, result)
}

func TestEventService_GetEventByID_EmptyID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	_, err := env.Service.GetEventByID(ctx, "")
	assert.ErrorIs(t, err, services2.ErrInvalidEventID)
}

func TestEventService_FindEvent_ByUserID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// Создаем события для разных пользователей
	events := []domain.Event{
		{
			Title:      "User1 Event1",
			StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			UserID:     userID1,
			OffsetTime: 0,
		},
		{
			Title:      "User1 Event2",
			StartDate:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
			UserID:     userID1,
			OffsetTime: 0,
		},
		{
			Title:      "User2 Event1",
			StartDate:  time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			UserID:     userID2,
			OffsetTime: 0,
		},
	}

	for _, e := range events {
		_, err := env.Service.CreateEvent(ctx, e)
		require.NoError(t, err)
	}

	// Ищем события пользователя 1
	result, err := env.Service.FindEvent(ctx, userID1, nil, nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, e := range result {
		assert.Equal(t, userID1, e.UserID)
	}
}

func TestEventService_FindEvent_ByDateRange(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Создаем события в разные даты
	events := []domain.Event{
		{
			Title:      "Event Jan 1",
			StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			UserID:     userID,
			OffsetTime: 0,
		},
		{
			Title:      "Event Jan 5",
			StartDate:  time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			UserID:     userID,
			OffsetTime: 0,
		},
		{
			Title:      "Event Jan 10",
			StartDate:  time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			UserID:     userID,
			OffsetTime: 0,
		},
	}

	for _, e := range events {
		_, err := env.Service.CreateEvent(ctx, e)
		require.NoError(t, err)
	}

	// Ищем события с 2 по 7 января
	startFrom := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	startTo := time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC)

	result, err := env.Service.FindEvent(ctx, userID, &startFrom, &startTo, nil, nil, nil)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Event Jan 5", result[0].Title)
}

func TestEventService_FindEvent_AllEvents(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	// Создаем события для разных пользователей
	for i := 0; i < 5; i++ {
		event := domain.Event{
			Title:      "Event " + string(rune('A'+i)),
			StartDate:  time.Date(2024, 1, i+1, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, i+1, 11, 0, 0, 0, time.UTC),
			UserID:     uuid.New().String(),
			OffsetTime: 0,
		}
		_, err := env.Service.CreateEvent(ctx, event)
		require.NoError(t, err)
	}

	// Получаем все события (без фильтров)
	result, err := env.Service.FindEvent(ctx, "", nil, nil, nil, nil, nil)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 5)
}

func TestEventService_TransactionRollback_OnError(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Создаем первое событие
	event1 := domain.Event{
		Title:      "Event 1",
		StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		UserID:     userID,
		OffsetTime: 0,
	}

	_, err := env.Service.CreateEvent(ctx, event1)
	require.NoError(t, err)

	// Пытаемся создать пересекающееся событие (должна быть ошибка)
	event2 := domain.Event{
		Title:      "Event 2",
		StartDate:  time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC),
		UserID:     userID,
		OffsetTime: 0,
	}

	_, err = env.Service.CreateEvent(ctx, event2)
	assert.ErrorIs(t, err, services2.ErrDateBusy)

	// Проверяем, что в БД только одно событие (транзакция откатилась)
	result, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Event 1", result[0].Title)
}

func TestEventService_NotificationSent_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create an event
	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test",
		UserID:      userID,
		OffsetTime:  0,
	}

	createdEvent, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Verify initially false
	retrieved, err := env.Service.GetEventByID(ctx, createdEvent.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.NotificationSent, "Initially notification should not be sent")

	// Mark notification as sent
	err = env.Service.NotificationSent(ctx, createdEvent.ID)
	require.NoError(t, err)

	// Verify the flag is set
	retrieved, err = env.Service.GetEventByID(ctx, createdEvent.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.NotificationSent, "Notification should be marked as sent")
}

func TestEventService_NotificationSent_EventNotFound(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	// Try to mark notification for non-existing event
	err := env.Service.NotificationSent(ctx, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, services2.ErrEventNotFound)
}

func TestEventService_NotificationSent_InvalidID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	// Try to mark notification with empty ID
	err := env.Service.NotificationSent(ctx, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, services2.ErrInvalidEventID)
}

func TestEventService_NotificationSent_Idempotent(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create an event
	event := domain.Event{
		Title:       "Idempotent Test",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test idempotency",
		UserID:      userID,
		OffsetTime:  0,
	}

	createdEvent, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Mark as sent twice
	err = env.Service.NotificationSent(ctx, createdEvent.ID)
	require.NoError(t, err)
	err = env.Service.NotificationSent(ctx, createdEvent.ID)
	require.NoError(t, err)

	// Verify still true
	retrieved, err := env.Service.GetEventByID(ctx, createdEvent.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.NotificationSent)
}

func TestEventService_NotificationSent_FilterEvents(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create multiple events
	event1 := domain.Event{
		Title:       "Event 1",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "First",
		UserID:      userID,
		OffsetTime:  0,
	}

	event2 := domain.Event{
		Title:       "Event 2",
		StartDate:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
		Description: "Second",
		UserID:      userID,
		OffsetTime:  0,
	}

	event3 := domain.Event{
		Title:       "Event 3",
		StartDate:   time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC),
		Description: "Third",
		UserID:      userID,
		OffsetTime:  0,
	}

	created1, err := env.Service.CreateEvent(ctx, event1)
	require.NoError(t, err)
	created2, err := env.Service.CreateEvent(ctx, event2)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, event3)
	require.NoError(t, err)

	// Mark first two events as sent
	err = env.Service.NotificationSent(ctx, created1.ID)
	require.NoError(t, err)
	err = env.Service.NotificationSent(ctx, created2.ID)
	require.NoError(t, err)

	// Filter by notificationSent = true
	notificationSentTrue := true
	sentEvents, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, &notificationSentTrue)
	require.NoError(t, err)
	assert.Len(t, sentEvents, 2, "Should find 2 events with notification sent")
	for _, e := range sentEvents {
		assert.True(t, e.NotificationSent)
	}

	// Filter by notificationSent = false
	notificationSentFalse := false
	notSentEvents, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, &notificationSentFalse)
	require.NoError(t, err)
	assert.Len(t, notSentEvents, 1, "Should find 1 event without notification sent")
	for _, e := range notSentEvents {
		assert.False(t, e.NotificationSent)
	}

	// No filter
	allEvents, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, allEvents, 3, "Should find all 3 events")
}

func TestEventService_DeleteOldEvents_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create events with different end dates
	oldEvent1 := domain.Event{
		Title:       "Old Event 1",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Old",
		UserID:      userID,
		OffsetTime:  0,
	}

	oldEvent2 := domain.Event{
		Title:       "Old Event 2",
		StartDate:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
		Description: "Old",
		UserID:      userID,
		OffsetTime:  0,
	}

	recentEvent := domain.Event{
		Title:       "Recent Event",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Recent",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, oldEvent1)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, oldEvent2)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, recentEvent)
	require.NoError(t, err)

	// Delete events with end_date <= 2024-01-03
	threshold := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	err = env.Service.DeleteOldEvents(ctx, &threshold)
	require.NoError(t, err)

	// Verify only recent event remains
	allEvents, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, allEvents, 1, "Should have 1 event remaining")
	assert.Equal(t, "Recent Event", allEvents[0].Title)
}

func TestEventService_DeleteOldEvents_NilThreshold(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	// Try to delete with nil threshold
	err := env.Service.DeleteOldEvents(ctx, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, services2.ErrInvalidThresholdDate)
}

func TestEventService_DeleteOldEvents_BoundaryCondition(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create event ending exactly at threshold
	event := domain.Event{
		Title:       "Boundary Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Description: "Boundary",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Delete with threshold equal to end_date
	threshold := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	err = env.Service.DeleteOldEvents(ctx, &threshold)
	require.NoError(t, err)

	// Verify event was deleted
	allEvents, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, allEvents, 0, "Event with end_date equal to threshold should be deleted")
}

func TestEventService_DeleteOldEvents_NoEventsToDelete(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create recent event
	event := domain.Event{
		Title:       "Recent Event",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Recent",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Delete with threshold before all events
	threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	err = env.Service.DeleteOldEvents(ctx, &threshold)
	require.NoError(t, err)

	// Verify event still exists
	allEvents, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, allEvents, 1, "Recent event should not be deleted")
}

func TestEventService_DeleteOldEvents_MultipleUsers(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// Create events for different users
	user1OldEvent := domain.Event{
		Title:       "User1 Old Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Old",
		UserID:      userID1,
		OffsetTime:  0,
	}

	user2OldEvent := domain.Event{
		Title:       "User2 Old Event",
		StartDate:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
		Description: "Old",
		UserID:      userID2,
		OffsetTime:  0,
	}

	user1RecentEvent := domain.Event{
		Title:       "User1 Recent Event",
		StartDate:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
		Description: "Recent",
		UserID:      userID1,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, user1OldEvent)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, user2OldEvent)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, user1RecentEvent)
	require.NoError(t, err)

	// Delete events with end_date <= 2024-01-05
	threshold := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	err = env.Service.DeleteOldEvents(ctx, &threshold)
	require.NoError(t, err)

	// Verify only recent event remains
	allEvents, err := env.Service.FindEvent(ctx, "", nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, allEvents, 1, "Should have 1 event remaining")
	assert.Equal(t, "User1 Recent Event", allEvents[0].Title)

	// Verify user1 has only recent event
	user1Events, err := env.Service.FindEvent(ctx, userID1, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, user1Events, 1, "User1 should have 1 event")

	// Verify user2 has no events
	user2Events, err := env.Service.FindEvent(ctx, userID2, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, user2Events, 0, "User2 should have 0 events")
}

func TestEventService_FindUpcomingEvent_Success(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create events with different start dates
	pastEvent := domain.Event{
		Title:       "Past Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Past",
		UserID:      userID,
		OffsetTime:  0,
	}

	upcomingEvent1 := domain.Event{
		Title:       "Upcoming Event 1",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Upcoming",
		UserID:      userID,
		OffsetTime:  0,
	}

	upcomingEvent2 := domain.Event{
		Title:       "Upcoming Event 2",
		StartDate:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
		Description: "Upcoming",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, pastEvent)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, upcomingEvent1)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, upcomingEvent2)
	require.NoError(t, err)

	// Find events starting from Jan 3
	threshold := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	events, err := env.Service.FindUpcomingEvent(ctx, userID, &threshold)
	require.NoError(t, err)

	// Should find only upcoming events
	assert.Len(t, events, 1)
	titles := make(map[string]bool)
	for _, e := range events {
		titles[e.Title] = true
	}
	assert.False(t, titles["Upcoming Event 1"])
	assert.False(t, titles["Upcoming Event 2"])
	assert.True(t, titles["Past Event"])
}

func TestEventService_FindUpcomingEvent_FilterByUser(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// Create events for different users
	user1Event := domain.Event{
		Title:       "User 1 Event",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "User 1",
		UserID:      userID1,
		OffsetTime:  0,
	}

	user2Event := domain.Event{
		Title:       "User 2 Event",
		StartDate:   time.Date(2024, 1, 5, 14, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 15, 0, 0, 0, time.UTC),
		Description: "User 2",
		UserID:      userID2,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, user1Event)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, user2Event)
	require.NoError(t, err)

	// Find upcoming events for user-1
	threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	events, err := env.Service.FindUpcomingEvent(ctx, userID1, &threshold)
	require.NoError(t, err)

	// Should find only user-1 event
	assert.Len(t, events, 0)
}

func TestEventService_FindUpcomingEvent_NilThreshold(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create events
	event1 := domain.Event{
		Title:       "Event 1",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "First",
		UserID:      userID,
		OffsetTime:  0,
	}

	event2 := domain.Event{
		Title:       "Event 2",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Second",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, event1)
	require.NoError(t, err)
	_, err = env.Service.CreateEvent(ctx, event2)
	require.NoError(t, err)

	// Find with nil threshold
	_, err = env.Service.FindUpcomingEvent(ctx, userID, nil)
	require.ErrorIs(t, err, services2.ErrInvalidThresholdDate)
}

func TestEventService_FindUpcomingEvent_BoundaryCondition(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create event starting exactly at threshold
	event := domain.Event{
		Title:       "Boundary Event",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Boundary",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Find with threshold equal to start date
	threshold := time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)
	events, err := env.Service.FindUpcomingEvent(ctx, userID, &threshold)
	require.NoError(t, err)

	// Should include event starting at threshold
	assert.Len(t, events, 1)
	assert.Equal(t, "Boundary Event", events[0].Title)
}

func TestEventService_FindUpcomingEvent_NoUpcomingEvents(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create only past events
	pastEvent := domain.Event{
		Title:       "Past Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Past",
		UserID:      userID,
		OffsetTime:  0,
	}

	_, err := env.Service.CreateEvent(ctx, pastEvent)
	require.NoError(t, err)

	// Find with threshold after all events
	threshold := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	events, err := env.Service.FindUpcomingEvent(ctx, userID, &threshold)
	require.NoError(t, err)

	// Should find no events
	assert.Len(t, events, 1)
}

func TestEventService_FindUpcomingEvent_WithOffsetTime(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create event with 1 hour offset (notification 1 hour before start)
	event := domain.Event{
		Title:       "Event with Offset",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Has offset",
		UserID:      userID,
		OffsetTime:  time.Hour,
	}

	_, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Threshold at 9:00 - event starts at 10:00 but notification should be at 9:00
	threshold := time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC)
	events, err := env.Service.FindUpcomingEvent(ctx, userID, &threshold)
	require.NoError(t, err)

	// Should find the event because start_date - offset_time (10:00 - 1h = 9:00) >= threshold (9:00)
	assert.Len(t, events, 1)
	assert.Equal(t, "Event with Offset", events[0].Title)
}

func TestEventService_FindUpcomingEvent_WithOffsetTime_BeforeNotification(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()
	userID := uuid.New().String()

	// Create event with 1 hour offset
	event := domain.Event{
		Title:       "Event with Offset",
		StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
		Description: "Has offset",
		UserID:      userID,
		OffsetTime:  time.Hour,
	}

	_, err := env.Service.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Threshold at 9:01 - notification time is 9:00, so event should not be found
	threshold := time.Date(2024, 1, 5, 9, 1, 0, 0, time.UTC)
	events, err := env.Service.FindUpcomingEvent(ctx, userID, &threshold)
	require.NoError(t, err)

	// Should not find the event because start_date - offset_time (9:00) < threshold (9:01)
	assert.Len(t, events, 1)
}
