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

func TestEventRepository_FindEvent(t *testing.T) {
	ctx := context.Background()
	crudRepo := NewEventCrudRepository()
	repo, err := NewEventRepository(crudRepo)
	require.NoError(t, err)

	// Event 1: Jan 1, 2024 10:00 - 11:00
	event1 := domain.Event{
		Title:       "Event 1",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "First event",
		UserID:      "user-1",
		OffsetTime:  0,
	}

	// Event 2: Jan 2, 2024 14:00 - 15:00
	event2 := domain.Event{
		Title:       "Event 2",
		StartDate:   time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC),
		Description: "Second event",
		UserID:      "user-1",
		OffsetTime:  0,
	}

	// Event 3: Jan 5, 2024 09:00 - 10:00
	event3 := domain.Event{
		Title:       "Event 3",
		StartDate:   time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		Description: "Third event",
		UserID:      "user-2",
		OffsetTime:  0,
	}

	// Event 4: Long event spanning multiple days (Jan 3 - Jan 6)
	event4 := domain.Event{
		Title:       "Long Event",
		StartDate:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 6, 23, 59, 59, 0, time.UTC),
		Description: "Multi-day event",
		UserID:      "user-1",
		OffsetTime:  0,
	}

	// Create all events
	_, err = repo.Create(ctx, nil, event1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, nil, event2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, nil, event3)
	require.NoError(t, err)
	_, err = repo.Create(ctx, nil, event4)
	require.NoError(t, err)

	t.Run("find events in date range", func(t *testing.T) {
		// Search for events from Jan 1 to Jan 2 (by start_date)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 2, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find event1 (Jan 1) and event2 (Jan 2)
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
	})

	t.Run("find events by start_date range", func(t *testing.T) {
		// Search for events starting from Jan 3 to Jan 5 (by start_date)
		from := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find event3 (Jan 5) and event4 (Jan 3)
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 3"])
		assert.True(t, titles["Long Event"])
	})

	t.Run("find no events outside range", func(t *testing.T) {
		// Search for events in Feb 2024
		from := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, events, 0)
	})

	t.Run("find all events in wide range", func(t *testing.T) {
		// Search for all events in January
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, events, 4)
	})

	t.Run("find events on exact date", func(t *testing.T) {
		// Search for events starting on Jan 2
		from := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 2, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find only event2 (starts on Jan 2)
		assert.Len(t, events, 1)
		if len(events) > 0 {
			assert.Equal(t, "Event 2", events[0].Title)
		}
	})

	t.Run("filter by userId", func(t *testing.T) {
		// Search for user-1 events in January
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "user-1", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find event1, event2, event4 (all belong to user-1)
		assert.Len(t, events, 3)
		for _, e := range events {
			assert.Equal(t, "user-1", e.UserID)
		}
	})

	t.Run("filter by userId - user-2", func(t *testing.T) {
		// Search for user-2 events in January
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "user-2", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find only event3 (belongs to user-2)
		assert.Len(t, events, 1)
		assert.Equal(t, "Event 3", events[0].Title)
		assert.Equal(t, "user-2", events[0].UserID)
	})

	t.Run("nil from and to parameters", func(t *testing.T) {
		// Search without date range (all events)
		events, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)

		// Should find all 4 events
		assert.Len(t, events, 4)
	})

	t.Run("filter by end_date range", func(t *testing.T) {
		// Search for events ending between Jan 1 11:00 and Jan 2 15:00
		endFrom := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, nil, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Should find event1 (ends Jan 1 11:00) and event2 (ends Jan 2 15:00)
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
	})

	t.Run("filter by both start_date and end_date", func(t *testing.T) {
		// Search for events: start_date >= Jan 1 AND end_date <= Jan 5 12:00
		startFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &startFrom, nil, nil, &endTo, nil)
		require.NoError(t, err)

		// Should find event1, event2, event3 (all end before Jan 5 12:00)
		// event4 (Long Event) ends Jan 6, so it should NOT be included
		assert.Len(t, events, 3)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
		assert.True(t, titles["Event 3"])
		assert.False(t, titles["Long Event"])
	})

	t.Run("filter by start and end date range - find long event", func(t *testing.T) {
		// Search for events: start_date <= Jan 3 AND end_date >= Jan 6
		startTo := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)
		endFrom := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, &startTo, &endFrom, nil, nil)
		require.NoError(t, err)

		// Should find only event4 (Long Event: Jan 3 - Jan 6)
		assert.Len(t, events, 1)
		assert.Equal(t, "Long Event", events[0].Title)
	})

	t.Run("filter events ending on specific day", func(t *testing.T) {
		// Search for events ending on Jan 5
		endFrom := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, nil, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Should find only event3 (ends Jan 5 10:00)
		assert.Len(t, events, 1)
		assert.Equal(t, "Event 3", events[0].Title)
	})

	t.Run("complex filter - start and end date ranges", func(t *testing.T) {
		// Search for events: start_date between Jan 1-3 AND end_date between Jan 1-3
		startFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		startTo := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		endFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &startFrom, &startTo, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Should find event1 (Jan 1 10:00-11:00) and event2 (Jan 2 14:00-15:00)
		// event3 starts Jan 5 - excluded
		// event4 (Long Event) ends Jan 6 - excluded
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
	})

	t.Run("find events within date range - overlapping events", func(t *testing.T) {
		// Ищем события, которые пересекаются с диапазоном Jan 2 - Jan 4
		// Логика: start_date <= Jan 4 AND end_date >= Jan 2
		startTo := time.Date(2024, 1, 4, 23, 59, 59, 0, time.UTC)
		endFrom := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, &startTo, &endFrom, nil, nil)
		require.NoError(t, err)

		// Должны найти:
		// - event2 (Jan 2 14:00-15:00) - полностью в диапазоне
		// - event4 (Long Event: Jan 3 - Jan 6) - начинается в диапазоне
		assert.GreaterOrEqual(t, len(events), 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 2"])
		assert.True(t, titles["Long Event"])
	})

	t.Run("find events fully contained in date range", func(t *testing.T) {
		// Ищем события, которые полностью находятся внутри Jan 1 - Jan 3
		// Логика: start_date >= Jan 1 AND end_date <= Jan 3
		startFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &startFrom, nil, nil, &endTo, nil)
		require.NoError(t, err)

		// Должны найти:
		// - event1 (Jan 1 10:00-11:00) - полностью внутри
		// - event2 (Jan 2 14:00-15:00) - полностью внутри
		// event3 начинается Jan 5 - исключен
		// event4 (Long Event) заканчивается Jan 6 - исключен
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
	})

	t.Run("find events that span across date range", func(t *testing.T) {
		// Ищем события, которые охватывают диапазон Jan 4 - Jan 5
		// Логика: start_date <= Jan 4 AND end_date >= Jan 5
		startTo := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
		endFrom := time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, &startTo, &endFrom, nil, nil)
		require.NoError(t, err)

		// Должны найти только event4 (Long Event: Jan 3 - Jan 6)
		// Он начинается до Jan 4 и заканчивается после Jan 5
		assert.Len(t, events, 1)
		assert.Equal(t, "Long Event", events[0].Title)
	})

	t.Run("find events starting in range", func(t *testing.T) {
		// Ищем события, которые начинаются в диапазоне Jan 1 - Jan 3
		startFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		startTo := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", &startFrom, &startTo, nil, nil, nil)
		require.NoError(t, err)

		// Должны найти:
		// - event1 (starts Jan 1)
		// - event2 (starts Jan 2)
		// - event4 (Long Event starts Jan 3)
		assert.Len(t, events, 3)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
		assert.True(t, titles["Long Event"])
	})

	t.Run("find events ending in range", func(t *testing.T) {
		// Ищем события, которые заканчиваются в диапазоне Jan 5 - Jan 7
		endFrom := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, nil, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Должны найти:
		// - event3 (ends Jan 5 10:00)
		// - event4 (Long Event ends Jan 6 23:59:59)
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 3"])
		assert.True(t, titles["Long Event"])
	})

	t.Run("find events with exact time overlap", func(t *testing.T) {
		// Ищем события, которые пересекаются с точным временем Jan 2 14:30
		// Логика: start_date <= Jan 2 14:30 AND end_date >= Jan 2 14:30
		targetTime := time.Date(2024, 1, 2, 14, 30, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, nil, "", nil, &targetTime, &targetTime, nil, nil)
		require.NoError(t, err)

		// Должны найти event2 (Jan 2 14:00-15:00)
		// 14:30 находится внутри этого события
		assert.GreaterOrEqual(t, len(events), 1)

		found := false
		for _, e := range events {
			if e.Title == "Event 2" {
				found = true
				break
			}
		}
		assert.True(t, found, "Event 2 should be found as it overlaps with target time")
	})

	t.Run("filter by notificationSent flag - true", func(t *testing.T) {
		// Create events with notificationSent flag
		eventWithNotification := domain.Event{
			Title:            "Event with notification",
			StartDate:        time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			EndDate:          time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC),
			Description:      "Notification sent",
			UserID:           "user-1",
			OffsetTime:       0,
			NotificationSent: true,
		}

		_, err := repo.Create(ctx, nil, eventWithNotification)
		require.NoError(t, err)

		// Search for events with notificationSent = true
		notificationSent := true
		events, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find only the event with notification sent
		assert.GreaterOrEqual(t, len(events), 1)
		for _, e := range events {
			assert.True(t, e.NotificationSent, "All events should have notificationSent = true")
		}
	})

	t.Run("filter by notificationSent flag - false", func(t *testing.T) {
		// Search for events with notificationSent = false
		notificationSent := false
		events, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find events without notification sent (event1, event2, event3, event4)
		assert.GreaterOrEqual(t, len(events), 4)
		for _, e := range events {
			assert.False(t, e.NotificationSent, "All events should have notificationSent = false")
		}
	})

	t.Run("filter by notificationSent with date range", func(t *testing.T) {
		// Search for events in January with notificationSent = false
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
		notificationSent := false

		events, err := repo.FindEvent(ctx, nil, "", &from, &to, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find event1, event2, event3, event4 (all in January with notificationSent = false)
		assert.Len(t, events, 4)
		for _, e := range events {
			assert.False(t, e.NotificationSent)
		}
	})

	t.Run("filter by notificationSent and userId", func(t *testing.T) {
		// Search for user-1 events with notificationSent = false
		notificationSent := false
		events, err := repo.FindEvent(ctx, nil, "user-1", nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find event1, event2, event4 (user-1 events with notificationSent = false)
		assert.GreaterOrEqual(t, len(events), 3)
		for _, e := range events {
			assert.Equal(t, "user-1", e.UserID)
			assert.False(t, e.NotificationSent)
		}
	})
}

func TestEventRepository_CRUD_Operations(t *testing.T) {
	ctx := context.Background()
	crudRepo := NewEventCrudRepository()
	repo, err := NewEventRepository(crudRepo)
	require.NoError(t, err)

	event := domain.Event{
		Title:       "Test Event",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Hour),
		Description: "Test Description",
		UserID:      "user-test",
		OffsetTime:  0,
	}

	var createdEventID string

	t.Run("create and retrieve", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)
		require.NotNil(t, createdEvent)
		assert.NotEmpty(t, createdEvent.ID, "ID should be auto-generated")
		createdEventID = createdEvent.ID
		retrieved, err := repo.GetByID(ctx, nil, createdEvent.ID)
		require.NoError(t, err)
		assert.Equal(t, createdEvent.ID, retrieved.ID)
		assert.Equal(t, event.Title, retrieved.Title)
	})

	t.Run("update event", func(t *testing.T) {
		event.Title = "Updated Title"
		updatedEvent, err := repo.Update(ctx, nil, createdEventID, event)
		require.NoError(t, err)
		require.NotNil(t, updatedEvent)

		retrieved, err := repo.GetByID(ctx, nil, createdEventID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)
	})

	t.Run("delete event", func(t *testing.T) {
		err := repo.Delete(ctx, nil, createdEventID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, nil, createdEventID)
		require.Error(t, err)
	})
}

func TestEventRepository_NotificationSent(t *testing.T) {
	ctx := context.Background()
	crudRepo := NewEventCrudRepository()
	repo, err := NewEventRepository(crudRepo)
	require.NoError(t, err)

	t.Run("mark notification as sent for existing event", func(t *testing.T) {
		// Create an event
		event := domain.Event{
			Title:       "Test Event",
			StartDate:   time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC),
			Description: "Test",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		createdEvent, err := repo.Create(ctx, nil, event)
		require.NoError(t, err)
		assert.False(t, createdEvent.NotificationSent, "Initially notification should not be sent")

		// Mark notification as sent
		err = repo.NotificationSent(ctx, nil, createdEvent.ID)
		require.NoError(t, err)

		// Verify the flag is set
		retrieved, err := repo.GetByID(ctx, nil, createdEvent.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.NotificationSent, "Notification should be marked as sent")
	})

	t.Run("mark notification as sent for non-existing event", func(t *testing.T) {
		err := repo.NotificationSent(ctx, nil, "non-existing-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
	})

	t.Run("filter events by notificationSent flag", func(t *testing.T) {
		// Clear repository
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create events
		event1 := domain.Event{
			Title:       "Event 1",
			StartDate:   time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC),
			Description: "First",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		event2 := domain.Event{
			Title:       "Event 2",
			StartDate:   time.Date(2024, 2, 2, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 2, 11, 0, 0, 0, time.UTC),
			Description: "Second",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		event3 := domain.Event{
			Title:       "Event 3",
			StartDate:   time.Date(2024, 2, 3, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 3, 11, 0, 0, 0, time.UTC),
			Description: "Third",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		created1, err := repo.Create(ctx, nil, event1)
		require.NoError(t, err)
		created2, err := repo.Create(ctx, nil, event2)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, event3)
		require.NoError(t, err)

		// Mark first two events as sent
		err = repo.NotificationSent(ctx, nil, created1.ID)
		require.NoError(t, err)
		err = repo.NotificationSent(ctx, nil, created2.ID)
		require.NoError(t, err)

		// Filter by notificationSent = true
		notificationSentTrue := true
		sentEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, &notificationSentTrue)
		require.NoError(t, err)
		assert.Len(t, sentEvents, 2, "Should find 2 events with notification sent")
		for _, e := range sentEvents {
			assert.True(t, e.NotificationSent)
		}

		// Filter by notificationSent = false
		notificationSentFalse := false
		notSentEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, &notificationSentFalse)
		require.NoError(t, err)
		assert.Len(t, notSentEvents, 1, "Should find 1 event without notification sent")
		for _, e := range notSentEvents {
			assert.False(t, e.NotificationSent)
		}

		// No filter
		allEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 3, "Should find all 3 events")
	})
}

func TestEventRepository_DeleteOldEvents(t *testing.T) {
	ctx := context.Background()
	crudRepo := NewEventCrudRepository()
	repo, err := NewEventRepository(crudRepo)
	require.NoError(t, err)

	t.Run("delete events with end_date before threshold", func(t *testing.T) {
		// Create events with different end dates
		oldEvent1 := domain.Event{
			Title:       "Old Event 1",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		oldEvent2 := domain.Event{
			Title:       "Old Event 2",
			StartDate:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		recentEvent := domain.Event{
			Title:       "Recent Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Recent",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, nil, oldEvent1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, oldEvent2)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, recentEvent)
		require.NoError(t, err)

		// Delete events with end_date <= 2024-01-03
		threshold := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, nil, &threshold)
		require.NoError(t, err)

		// Verify only recent event remains
		allEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 1, "Should have 1 event remaining")
		assert.Equal(t, "Recent Event", allEvents[0].Title)
	})

	t.Run("delete events with end_date equal to threshold", func(t *testing.T) {
		// Clear repository
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create event ending exactly at threshold
		event := domain.Event{
			Title:       "Boundary Event",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Description: "Boundary",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, nil, event)
		require.NoError(t, err)

		// Delete with threshold equal to end_date
		threshold := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, nil, &threshold)
		require.NoError(t, err)

		// Verify event was deleted
		allEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 0, "Event with end_date equal to threshold should be deleted")
	})

	t.Run("no events to delete", func(t *testing.T) {
		// Clear repository
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create recent event
		event := domain.Event{
			Title:       "Recent Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Recent",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, nil, event)
		require.NoError(t, err)

		// Delete with threshold before all events
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, nil, &threshold)
		require.NoError(t, err)

		// Verify event still exists
		allEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 1, "Recent event should not be deleted")
	})

	t.Run("delete all events", func(t *testing.T) {
		// Clear repository
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create old events
		event1 := domain.Event{
			Title:       "Event 1",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		event2 := domain.Event{
			Title:       "Event 2",
			StartDate:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, nil, event1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, event2)
		require.NoError(t, err)

		// Delete with threshold after all events
		threshold := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, nil, &threshold)
		require.NoError(t, err)

		// Verify all events deleted
		allEvents, err := repo.FindEvent(ctx, nil, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 0, "All events should be deleted")
	})
}

func TestEventRepository_FindUpcomingEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("find upcoming events after threshold", func(t *testing.T) {
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create events with different start dates
		pastEvent := domain.Event{
			Title:       "Past Event",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Past",
			UserID:      "user-1",
			OffsetTime:  900000000,
		}

		upcomingEvent1 := domain.Event{
			Title:       "Upcoming Event 1",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Upcoming",
			UserID:      "user-1",
			OffsetTime:  900000000,
		}

		upcomingEvent2 := domain.Event{
			Title:       "Upcoming Event 2",
			StartDate:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			Description: "Upcoming",
			UserID:      "user-1",
			OffsetTime:  900000000,
		}

		_, err = repo.Create(ctx, nil, pastEvent)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, upcomingEvent1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, upcomingEvent2)
		require.NoError(t, err)

		// Find events starting from Jan 3
		threshold := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, nil, "", &threshold)
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
	})

	t.Run("filter by user ID", func(t *testing.T) {
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create events for different users
		user1Event := domain.Event{
			Title:       "User 1 Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "User 1",
			UserID:      "user-1",
			OffsetTime:  900000000,
		}

		user2Event := domain.Event{
			Title:       "User 2 Event",
			StartDate:   time.Date(2024, 1, 5, 14, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 15, 0, 0, 0, time.UTC),
			Description: "User 2",
			UserID:      "user-2",
			OffsetTime:  900000000,
		}

		_, err = repo.Create(ctx, nil, user1Event)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, user2Event)
		require.NoError(t, err)

		// Find upcoming events for user-1
		threshold := time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, nil, "user-1", &threshold)
		require.NoError(t, err)

		// Should find only user-1 event
		assert.Len(t, events, 1)
		assert.Equal(t, "User 1 Event", events[0].Title)
		assert.Equal(t, "user-1", events[0].UserID)
	})

	t.Run("nil threshold returns all events", func(t *testing.T) {
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create events
		event1 := domain.Event{
			Title:       "Event 1",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "First",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		event2 := domain.Event{
			Title:       "Event 2",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Second",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, nil, event1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, nil, event2)
		require.NoError(t, err)

		// Find with nil threshold
		events, err := repo.FindUpcomingEvent(ctx, nil, "", nil)
		require.NoError(t, err)

		// Should find all events
		assert.Len(t, events, 2)
	})

	t.Run("boundary condition - exact threshold time", func(t *testing.T) {
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create event starting exactly at threshold
		event := domain.Event{
			Title:       "Boundary Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Boundary",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, nil, event)
		require.NoError(t, err)

		// Find with threshold equal to start date
		threshold := time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, nil, "", &threshold)
		require.NoError(t, err)

		// Should include event starting at threshold
		assert.Len(t, events, 1)
		assert.Equal(t, "Boundary Event", events[0].Title)
	})

	t.Run("no upcoming events", func(t *testing.T) {
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Create only past events
		pastEvent := domain.Event{
			Title:       "Past Event",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Past",
			UserID:      "user-1",
			OffsetTime:  0,
		}

		createdEvent, err := repo.Create(ctx, nil, pastEvent)
		require.NoError(t, err)
		err = repo.NotificationSent(ctx, nil, createdEvent.ID)
		require.NoError(t, err)
		// Find with threshold after all events
		threshold := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, nil, "", &threshold)
		require.NoError(t, err)

		// Should find no events
		assert.Len(t, events, 0)
	})

	t.Run("empty repository", func(t *testing.T) {
		crudRepo := NewEventCrudRepository()
		repo, err := NewEventRepository(crudRepo)
		require.NoError(t, err)

		// Find in empty repository
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, nil, "", &threshold)
		require.NoError(t, err)

		// Should return empty slice
		assert.Len(t, events, 0)
	})
}
