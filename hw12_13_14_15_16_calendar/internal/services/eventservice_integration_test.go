//go:build integration
// +build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
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
			expectedErr: ErrInvalidEventTitle,
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
			expectedErr: ErrInvalidUserID,
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
			expectedErr: ErrInvalidStartDate,
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
			expectedErr: ErrInvalidEndDate,
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
			expectedErr: ErrInvalidDateRange,
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
	assert.ErrorIs(t, err, ErrDateBusy)
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
	assert.ErrorIs(t, err, ErrEventNotFound)
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
	assert.ErrorIs(t, err, ErrInvalidEventID)
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
	assert.ErrorIs(t, err, ErrEventNotFound)
	assert.Nil(t, result)
}

func TestEventService_DeleteEvent_EmptyID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	err := env.Service.DeleteEvent(ctx, "")
	assert.ErrorIs(t, err, ErrInvalidEventID)
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

	assert.ErrorIs(t, err, ErrEventNotFound)
	assert.Nil(t, result)
}

func TestEventService_GetEventByID_EmptyID(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.CleanupTestData(t)

	ctx := context.Background()

	_, err := env.Service.GetEventByID(ctx, "")
	assert.ErrorIs(t, err, ErrInvalidEventID)
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
	result, err := env.Service.FindEvent(ctx, userID1, nil, nil, nil, nil)

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

	result, err := env.Service.FindEvent(ctx, userID, &startFrom, &startTo, nil, nil)

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
	result, err := env.Service.FindEvent(ctx, "", nil, nil, nil, nil)

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
	assert.ErrorIs(t, err, ErrDateBusy)

	// Проверяем, что в БД только одно событие (транзакция откатилась)
	result, err := env.Service.FindEvent(ctx, userID, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Event 1", result[0].Title)
}
