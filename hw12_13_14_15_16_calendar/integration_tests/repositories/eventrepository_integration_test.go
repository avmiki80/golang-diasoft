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

func TestEventRepository_FindEvent_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	// Event 1: Jan 1, 2024 10:00 - 11:00
	event1 := domain.Event{
		Title:       "Event 1",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "First event",
		UserID:      "550e8400-e29b-41d4-a716-446655440001",
		OffsetTime:  0,
	}

	// Event 2: Jan 2, 2024 14:00 - 15:00
	event2 := domain.Event{
		Title:       "Event 2",
		StartDate:   time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC),
		Description: "Second event",
		UserID:      "550e8400-e29b-41d4-a716-446655440001",
		OffsetTime:  0,
	}

	// Event 3: Jan 5, 2024 09:00 - 10:00
	event3 := domain.Event{
		Title:       "Event 3",
		StartDate:   time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		Description: "Third event",
		UserID:      "550e8400-e29b-41d4-a716-446655440002",
		OffsetTime:  0,
	}

	// Event 4: Long event spanning multiple days (Jan 3 - Jan 6)
	event4 := domain.Event{
		Title:       "Long Event",
		StartDate:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 6, 23, 59, 59, 0, time.UTC),
		Description: "Multi-day event",
		UserID:      "550e8400-e29b-41d4-a716-446655440001",
		OffsetTime:  0,
	}

	// Create all events
	_, err = repo.Create(ctx, db, event1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, event2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, event3)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, event4)
	require.NoError(t, err)

	t.Run("find events in date range", func(t *testing.T) {
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 2, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find Event 1 (Jan 1) and Event 2 (Jan 2) by start_date
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 1"])
		assert.True(t, titles["Event 2"])
	})

	t.Run("find events by start_date range", func(t *testing.T) {
		from := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find Event 3 (Jan 5) and Long Event (Jan 3) by start_date
		assert.Len(t, events, 2)

		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Event 3"])
		assert.True(t, titles["Long Event"])
	})

	t.Run("find no events outside range", func(t *testing.T) {
		from := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, events, 0)
	})

	t.Run("find all events in wide range", func(t *testing.T) {
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, events, 4)
	})

	t.Run("events are ordered by start_date", func(t *testing.T) {
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		require.Len(t, events, 4)

		for i := 1; i < len(events); i++ {
			assert.True(t, events[i-1].StartDate.Before(events[i].StartDate) ||
				events[i-1].StartDate.Equal(events[i].StartDate),
				"Events should be ordered by start_date")
		}
	})

	t.Run("filter by userId", func(t *testing.T) {
		// Search for user-1 events in January
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
		userID := "550e8400-e29b-41d4-a716-446655440001"

		events, err := repo.FindEvent(ctx, db, userID, &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find Event 1, Event 2, Long Event (all belong to user-1)
		assert.Len(t, events, 3)
		for _, e := range events {
			assert.Equal(t, userID, e.UserID)
		}
	})

	t.Run("filter by userId - user-2", func(t *testing.T) {
		// Search for user-2 events in January
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
		userID := "550e8400-e29b-41d4-a716-446655440002"

		events, err := repo.FindEvent(ctx, db, userID, &from, &to, nil, nil, nil)
		require.NoError(t, err)

		// Should find only Event 3 (belongs to user-2)
		assert.Len(t, events, 1)
		assert.Equal(t, "Event 3", events[0].Title)
		assert.Equal(t, userID, events[0].UserID)
	})

	t.Run("nil from and to parameters", func(t *testing.T) {
		// Search without date range (all events)
		events, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)

		// Should find all 4 events
		assert.Len(t, events, 4)
	})

	t.Run("filter by end_date range", func(t *testing.T) {
		// Search for events ending between Jan 1 11:00 and Jan 2 15:00
		endFrom := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", nil, nil, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Should find Event 1 (ends Jan 1 11:00) and Event 2 (ends Jan 2 15:00)
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

		events, err := repo.FindEvent(ctx, db, "", &startFrom, nil, nil, &endTo, nil)
		require.NoError(t, err)

		// Should find Event 1, Event 2, Event 3 (all end before Jan 5 12:00)
		// Long Event ends Jan 6, so it should NOT be included
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

		events, err := repo.FindEvent(ctx, db, "", nil, &startTo, &endFrom, nil, nil)
		require.NoError(t, err)

		// Should find only Long Event (starts Jan 3, ends Jan 6)
		assert.Len(t, events, 1)
		assert.Equal(t, "Long Event", events[0].Title)
	})

	t.Run("filter events ending on specific day", func(t *testing.T) {
		// Search for events ending on Jan 5
		endFrom := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", nil, nil, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Should find only Event 3 (ends Jan 5 10:00)
		assert.Len(t, events, 1)
		assert.Equal(t, "Event 3", events[0].Title)
	})

	t.Run("complex filter - start and end date ranges", func(t *testing.T) {
		// Search for events: start_date between Jan 1-3 AND end_date between Jan 1-3
		startFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		startTo := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		endFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &startFrom, &startTo, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Should find Event 1 (Jan 1 10:00-11:00) and Event 2 (Jan 2 14:00-15:00)
		// Event 3 starts Jan 5 - excluded
		// Long Event ends Jan 6 - excluded
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

		events, err := repo.FindEvent(ctx, db, "", nil, &startTo, &endFrom, nil, nil)
		require.NoError(t, err)

		// Должны найти:
		// - Event 2 (Jan 2 14:00-15:00) - полностью в диапазоне
		// - Long Event (Jan 3 - Jan 6) - начинается в диапазоне
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

		events, err := repo.FindEvent(ctx, db, "", &startFrom, nil, nil, &endTo, nil)
		require.NoError(t, err)

		// Должны найти:
		// - Event 1 (Jan 1 10:00-11:00) - полностью внутри
		// - Event 2 (Jan 2 14:00-15:00) - полностью внутри
		// Event 3 начинается Jan 5 - исключен
		// Long Event заканчивается Jan 6 - исключен
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

		events, err := repo.FindEvent(ctx, db, "", nil, &startTo, &endFrom, nil, nil)
		require.NoError(t, err)

		// Должны найти только Long Event (Jan 3 - Jan 6)
		// Он начинается до Jan 4 и заканчивается после Jan 5
		assert.Len(t, events, 1)
		assert.Equal(t, "Long Event", events[0].Title)
	})

	t.Run("find events starting in range", func(t *testing.T) {
		// Ищем события, которые начинаются в диапазоне Jan 1 - Jan 3
		startFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		startTo := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &startFrom, &startTo, nil, nil, nil)
		require.NoError(t, err)

		// Должны найти:
		// - Event 1 (starts Jan 1)
		// - Event 2 (starts Jan 2)
		// - Long Event (starts Jan 3)
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

		events, err := repo.FindEvent(ctx, db, "", nil, nil, &endFrom, &endTo, nil)
		require.NoError(t, err)

		// Должны найти:
		// - Event 3 (ends Jan 5 10:00)
		// - Long Event (ends Jan 6 23:59:59)
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

		events, err := repo.FindEvent(ctx, db, "", nil, &targetTime, &targetTime, nil, nil)
		require.NoError(t, err)

		// Должны найти Event 2 (Jan 2 14:00-15:00)
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
}

func TestEventRepository_CRUD_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	event := domain.Event{
		Title:       "Test Event CRUD",
		StartDate:   time.Now().Truncate(time.Second),
		EndDate:     time.Now().Add(time.Hour).Truncate(time.Second),
		Description: "Test Description",
		UserID:      "550e8400-e29b-41d4-a716-446655440099",
		OffsetTime:  0,
	}

	var createdEventID string

	t.Run("create and retrieve", func(t *testing.T) {
		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)
		require.NotNil(t, createdEvent)
		assert.NotEmpty(t, createdEvent.ID, "ID should be auto-generated")

		createdEventID = createdEvent.ID

		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.Equal(t, createdEvent.ID, retrieved.ID)
		assert.Equal(t, event.Title, retrieved.Title)
	})

	t.Run("update event", func(t *testing.T) {
		event.Title = "Updated Title CRUD"
		updatedEvent, err := repo.Update(ctx, db, createdEventID, event)
		require.NoError(t, err)
		require.NotNil(t, updatedEvent)

		retrieved, err := repo.GetByID(ctx, db, createdEventID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title CRUD", retrieved.Title)
	})

	t.Run("delete event", func(t *testing.T) {
		err := repo.Delete(ctx, db, createdEventID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, db, createdEventID)
		require.Error(t, err)
	})
}

func TestEventRepository_BoundaryConditions_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	t.Run("event starting exactly at 'from' boundary", func(t *testing.T) {
		event := domain.Event{
			Title:       "Boundary Event Start",
			StartDate:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC),
			Description: "Starts at boundary",
			UserID:      "550e8400-e29b-41d4-a716-446655440088",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		from := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)

		found := false
		for _, e := range events {
			if e.Title == "Boundary Event Start" {
				found = true
				break
			}
		}
		assert.True(t, found, "Event starting at 'from' boundary should be found")
	})

	t.Run("event ending exactly at 'to' boundary", func(t *testing.T) {
		cleanupTestData(t, db)

		event := domain.Event{
			Title:       "Boundary Event End",
			StartDate:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			Description: "Ends at boundary",
			UserID:      "550e8400-e29b-41d4-a716-446655440088",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		from := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)

		found := false
		for _, e := range events {
			if e.Title == "Boundary Event End" {
				found = true
				break
			}
		}
		assert.True(t, found, "Event ending at 'to' boundary should be found")
	})
}

func TestEventRepository_NotificationSent_Method_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	t.Run("mark notification as sent for existing event", func(t *testing.T) {
		// Create an event
		event := domain.Event{
			Title:       "Test Event",
			StartDate:   time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC),
			Description: "Test",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Verify initially false
		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.False(t, retrieved.NotificationSent, "Initially notification should not be sent")

		// Mark notification as sent
		err = repo.NotificationSent(ctx, db, createdEvent.ID)
		require.NoError(t, err)

		// Verify the flag is set
		retrieved, err = repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.NotificationSent, "Notification should be marked as sent")
	})

	t.Run("mark notification as sent for non-existing event", func(t *testing.T) {
		err := repo.NotificationSent(ctx, db, "550e8400-e29b-41d4-a716-446655440002")
		require.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrEntityNotFound)
	})

	t.Run("mark notification as sent is idempotent", func(t *testing.T) {
		// Create an event
		event := domain.Event{
			Title:       "Idempotent Test",
			StartDate:   time.Date(2024, 2, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 5, 11, 0, 0, 0, time.UTC),
			Description: "Test idempotency",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		createdEvent, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Mark as sent twice
		err = repo.NotificationSent(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		err = repo.NotificationSent(ctx, db, createdEvent.ID)
		require.NoError(t, err)

		// Verify still true
		retrieved, err := repo.GetByID(ctx, db, createdEvent.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.NotificationSent)
	})
}

func TestEventRepository_NotificationSent_Filter_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	// Event with notification sent
	eventWithNotification := domain.Event{
		Title:       "Event with notification",
		StartDate:   time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC),
		Description: "Notification sent",
		UserID:      "550e8400-e29b-41d4-a716-446655440001",
		OffsetTime:  0,
	}

	// Event without notification
	eventWithoutNotification := domain.Event{
		Title:       "Event without notification",
		StartDate:   time.Date(2024, 2, 2, 14, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 2, 2, 15, 0, 0, 0, time.UTC),
		Description: "No notification",
		UserID:      "550e8400-e29b-41d4-a716-446655440001",
		OffsetTime:  0,
	}

	// Another event without notification
	eventWithoutNotification2 := domain.Event{
		Title:       "Another event without notification",
		StartDate:   time.Date(2024, 2, 3, 9, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 2, 3, 10, 0, 0, 0, time.UTC),
		Description: "No notification",
		UserID:      "550e8400-e29b-41d4-a716-446655440002",
		OffsetTime:  0,
	}

	// Create all events
	createdEvent, err := repo.Create(ctx, db, eventWithNotification)
	require.NoError(t, err)
	err = repo.NotificationSent(ctx, db, createdEvent.ID)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, eventWithoutNotification)
	require.NoError(t, err)
	_, err = repo.Create(ctx, db, eventWithoutNotification2)
	require.NoError(t, err)

	t.Run("filter by notificationSent = true", func(t *testing.T) {
		notificationSent := true
		events, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find only the event with notification sent
		assert.GreaterOrEqual(t, len(events), 1)
		for _, e := range events {
			assert.True(t, e.NotificationSent, "All events should have notificationSent = true")
		}

		// Check that our specific event is found
		found := false
		for _, e := range events {
			if e.Title == "Event with notification" {
				found = true
				break
			}
		}
		assert.True(t, found, "Event with notification should be found")
	})

	t.Run("filter by notificationSent = false", func(t *testing.T) {
		notificationSent := false
		events, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find events without notification sent
		assert.GreaterOrEqual(t, len(events), 2)
		for _, e := range events {
			assert.False(t, e.NotificationSent, "All events should have notificationSent = false")
		}
	})

	t.Run("filter by notificationSent with date range", func(t *testing.T) {
		from := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC)
		notificationSent := false

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find events in February with notificationSent = false
		assert.GreaterOrEqual(t, len(events), 2)
		for _, e := range events {
			assert.False(t, e.NotificationSent)
			assert.True(t, e.StartDate.After(from) || e.StartDate.Equal(from))
			assert.True(t, e.StartDate.Before(to) || e.StartDate.Equal(to))
		}
	})

	t.Run("filter by notificationSent and userId", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440001"
		notificationSent := false

		events, err := repo.FindEvent(ctx, db, userID, nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find user-1 events with notificationSent = false
		assert.GreaterOrEqual(t, len(events), 1)
		for _, e := range events {
			assert.Equal(t, userID, e.UserID)
			assert.False(t, e.NotificationSent)
		}

		// Check that our specific event is found
		found := false
		for _, e := range events {
			if e.Title == "Event without notification" {
				found = true
				break
			}
		}
		assert.True(t, found, "Event without notification for user-1 should be found")
	})

	t.Run("filter by notificationSent = true and userId", func(t *testing.T) {
		userID := "550e8400-e29b-41d4-a716-446655440001"
		notificationSent := true

		events, err := repo.FindEvent(ctx, db, userID, nil, nil, nil, nil, &notificationSent)
		require.NoError(t, err)

		// Should find user-1 events with notificationSent = true
		assert.GreaterOrEqual(t, len(events), 1)
		for _, e := range events {
			assert.Equal(t, userID, e.UserID)
			assert.True(t, e.NotificationSent)
		}

		// Check that our specific event is found
		found := false
		for _, e := range events {
			if e.Title == "Event with notification" {
				found = true
				break
			}
		}
		assert.True(t, found, "Event with notification for user-1 should be found")
	})

	t.Run("no filter on notificationSent (nil)", func(t *testing.T) {
		events, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)

		// Should find all events regardless of notification status
		assert.GreaterOrEqual(t, len(events), 3)
	})
}

func TestEventRepository_DeleteOldEvents_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	t.Run("delete events with end_date before threshold", func(t *testing.T) {
		// Create events with different end dates
		oldEvent1 := domain.Event{
			Title:       "Old Event 1",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		oldEvent2 := domain.Event{
			Title:       "Old Event 2",
			StartDate:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		recentEvent := domain.Event{
			Title:       "Recent Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Recent",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, oldEvent1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, oldEvent2)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, recentEvent)
		require.NoError(t, err)

		// Delete events with end_date <= 2024-01-03
		threshold := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, db, &threshold)
		require.NoError(t, err)

		// Verify only recent event remains
		allEvents, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 1, "Should have 1 event remaining")
		assert.Equal(t, "Recent Event", allEvents[0].Title)
	})

	t.Run("delete events with end_date equal to threshold", func(t *testing.T) {
		// Clean up previous test data
		cleanupTestData(t, db)

		// Create event ending exactly at threshold
		event := domain.Event{
			Title:       "Boundary Event",
			StartDate:   time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
			Description: "Boundary",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Delete with threshold equal to end_date
		threshold := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, db, &threshold)
		require.NoError(t, err)

		// Verify event was deleted
		allEvents, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 0, "Event with end_date equal to threshold should be deleted")
	})

	t.Run("no events to delete", func(t *testing.T) {
		// Clean up previous test data
		cleanupTestData(t, db)

		// Create recent event
		event := domain.Event{
			Title:       "Recent Event",
			StartDate:   time.Date(2024, 3, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 3, 5, 11, 0, 0, 0, time.UTC),
			Description: "Recent",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Delete with threshold before all events
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, db, &threshold)
		require.NoError(t, err)

		// Verify event still exists
		allEvents, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 1, "Recent event should not be deleted")
	})

	t.Run("delete all events", func(t *testing.T) {
		// Clean up previous test data
		cleanupTestData(t, db)

		// Create old events
		event1 := domain.Event{
			Title:       "Event 1",
			StartDate:   time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 4, 1, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		event2 := domain.Event{
			Title:       "Event 2",
			StartDate:   time.Date(2024, 4, 2, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 4, 2, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, event1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, event2)
		require.NoError(t, err)

		// Delete with threshold after all events
		threshold := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, db, &threshold)
		require.NoError(t, err)

		// Verify all events deleted
		allEvents, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 0, "All events should be deleted")
	})

	t.Run("delete multiple old events from different users", func(t *testing.T) {
		// Clean up previous test data
		cleanupTestData(t, db)

		// Create events for different users
		user1Event1 := domain.Event{
			Title:       "User1 Old Event",
			StartDate:   time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 5, 1, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		user2Event1 := domain.Event{
			Title:       "User2 Old Event",
			StartDate:   time.Date(2024, 5, 2, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 5, 2, 11, 0, 0, 0, time.UTC),
			Description: "Old",
			UserID:      "550e8400-e29b-41d4-a716-446655440002",
			OffsetTime:  0,
		}

		user1Event2 := domain.Event{
			Title:       "User1 Recent Event",
			StartDate:   time.Date(2024, 5, 10, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 5, 10, 11, 0, 0, 0, time.UTC),
			Description: "Recent",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err := repo.Create(ctx, db, user1Event1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, user2Event1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, user1Event2)
		require.NoError(t, err)

		// Delete events with end_date <= 2024-05-05
		threshold := time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC)
		err = repo.DeleteOldEvents(ctx, db, &threshold)
		require.NoError(t, err)

		// Verify only recent event remains
		allEvents, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, allEvents, 1, "Should have 1 event remaining")
		assert.Equal(t, "User1 Recent Event", allEvents[0].Title)
	})
}

func TestEventRepository_FindUpcomingEvent_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := db2.NewEventCrudRepository(db)
	repo, err := db2.NewEventRepository(crudRepo)
	require.NoError(t, err)

	t.Run("find upcoming events after threshold", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create events with different start dates
		pastEvent := domain.Event{
			Title:       "Past Event",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Past",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		upcomingEvent1 := domain.Event{
			Title:       "Upcoming Event 1",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Upcoming",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		upcomingEvent2 := domain.Event{
			Title:       "Upcoming Event 2",
			StartDate:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			Description: "Upcoming",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, db, pastEvent)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, upcomingEvent1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, upcomingEvent2)
		require.NoError(t, err)

		// Find events starting from Jan 3
		threshold := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "", &threshold)
		require.NoError(t, err)

		// Should find only upcoming events
		assert.Len(t, events, 1)
		titles := make(map[string]bool)
		for _, e := range events {
			titles[e.Title] = true
		}
		assert.True(t, titles["Past Event"])
	})

	t.Run("filter by user ID", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create events for different users
		user1Event := domain.Event{
			Title:       "User 1 Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "User 1",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		user2Event := domain.Event{
			Title:       "User 2 Event",
			StartDate:   time.Date(2024, 1, 5, 14, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 15, 0, 0, 0, time.UTC),
			Description: "User 2",
			UserID:      "550e8400-e29b-41d4-a716-446655440002",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, db, user1Event)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, user2Event)
		require.NoError(t, err)

		// Find upcoming events for user-1
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "550e8400-e29b-41d4-a716-446655440001", &threshold)
		require.NoError(t, err)

		// Should find only user-1 event
		assert.Len(t, events, 0)
	})

	t.Run("nil threshold returns all events", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create events
		event1 := domain.Event{
			Title:       "Event 1",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "First",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		event2 := domain.Event{
			Title:       "Event 2",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Second",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, db, event1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, db, event2)
		require.NoError(t, err)

		// Find with nil threshold
		events, err := repo.FindUpcomingEvent(ctx, db, "", nil)
		require.NoError(t, err)

		// Should find all events
		assert.Len(t, events, 2)
	})

	t.Run("boundary condition - exact threshold time", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create event starting exactly at threshold
		event := domain.Event{
			Title:       "Boundary Event",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Boundary",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Find with threshold equal to start date
		threshold := time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "", &threshold)
		require.NoError(t, err)

		// Should include event starting at threshold
		assert.Len(t, events, 1)
		assert.Equal(t, "Boundary Event", events[0].Title)
	})

	t.Run("no upcoming events", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create only past events
		pastEvent := domain.Event{
			Title:       "Past Event",
			StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			Description: "Past",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  0,
		}

		_, err = repo.Create(ctx, db, pastEvent)
		require.NoError(t, err)

		// Find with threshold after all events
		threshold := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "", &threshold)
		require.NoError(t, err)

		// Should find no events
		assert.Len(t, events, 1)
	})

	t.Run("empty repository", func(t *testing.T) {
		cleanupTestData(t, db)

		// Find in empty repository
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "", &threshold)
		require.NoError(t, err)

		// Should return empty slice
		assert.Len(t, events, 0)
	})

	t.Run("with offset_time - event notification time", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create event with 1 hour offset (notification 1 hour before start)
		event := domain.Event{
			Title:       "Event with Offset",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Has offset",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  time.Hour,
		}

		_, err = repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Threshold at 9:00 - event starts at 10:00 but notification should be at 9:00
		threshold := time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "", &threshold)
		require.NoError(t, err)

		// Should find the event because start_date - offset_time (10:00 - 1h = 9:00) >= threshold (9:00)
		assert.Len(t, events, 1)
		assert.Equal(t, "Event with Offset", events[0].Title)
	})

	t.Run("with offset_time - before notification time", func(t *testing.T) {
		cleanupTestData(t, db)

		// Create event with 1 hour offset
		event := domain.Event{
			Title:       "Event with Offset",
			StartDate:   time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			Description: "Has offset",
			UserID:      "550e8400-e29b-41d4-a716-446655440001",
			OffsetTime:  time.Hour,
		}

		_, err = repo.Create(ctx, db, event)
		require.NoError(t, err)

		// Threshold at 9:01 - notification time is 9:00, so event should not be found
		threshold := time.Date(2024, 1, 5, 9, 1, 0, 0, time.UTC)
		events, err := repo.FindUpcomingEvent(ctx, db, "", &threshold)
		require.NoError(t, err)

		assert.Len(t, events, 1)
	})
}
