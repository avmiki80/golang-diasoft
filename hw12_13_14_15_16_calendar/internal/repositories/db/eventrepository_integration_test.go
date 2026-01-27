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

func TestEventRepository_FindEvent_WithTestcontainers(t *testing.T) {
	_, db := SetupPostgresContainer(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	crudRepo := NewEventCrudRepository(db)
	repo, err := NewEventRepository(crudRepo)
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

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
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

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
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

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
		require.NoError(t, err)
		assert.Len(t, events, 0)
	})

	t.Run("find all events in wide range", func(t *testing.T) {
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
		require.NoError(t, err)
		assert.Len(t, events, 4)
	})

	t.Run("events are ordered by start_date", func(t *testing.T) {
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
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

		events, err := repo.FindEvent(ctx, db, userID, &from, &to, nil, nil)
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

		events, err := repo.FindEvent(ctx, db, userID, &from, &to, nil, nil)
		require.NoError(t, err)

		// Should find only Event 3 (belongs to user-2)
		assert.Len(t, events, 1)
		assert.Equal(t, "Event 3", events[0].Title)
		assert.Equal(t, userID, events[0].UserID)
	})

	t.Run("nil from and to parameters", func(t *testing.T) {
		// Search without date range (all events)
		events, err := repo.FindEvent(ctx, db, "", nil, nil, nil, nil)
		require.NoError(t, err)

		// Should find all 4 events
		assert.Len(t, events, 4)
	})

	t.Run("filter by end_date range", func(t *testing.T) {
		// Search for events ending between Jan 1 11:00 and Jan 2 15:00
		endFrom := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", nil, nil, &endFrom, &endTo)
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

		events, err := repo.FindEvent(ctx, db, "", &startFrom, nil, nil, &endTo)
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

		events, err := repo.FindEvent(ctx, db, "", nil, &startTo, &endFrom, nil)
		require.NoError(t, err)

		// Should find only Long Event (starts Jan 3, ends Jan 6)
		assert.Len(t, events, 1)
		assert.Equal(t, "Long Event", events[0].Title)
	})

	t.Run("filter events ending on specific day", func(t *testing.T) {
		// Search for events ending on Jan 5
		endFrom := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
		endTo := time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC)

		events, err := repo.FindEvent(ctx, db, "", nil, nil, &endFrom, &endTo)
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

		events, err := repo.FindEvent(ctx, db, "", &startFrom, &startTo, &endFrom, &endTo)
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

		events, err := repo.FindEvent(ctx, db, "", nil, &startTo, &endFrom, nil)
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

		events, err := repo.FindEvent(ctx, db, "", &startFrom, nil, nil, &endTo)
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

		events, err := repo.FindEvent(ctx, db, "", nil, &startTo, &endFrom, nil)
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

		events, err := repo.FindEvent(ctx, db, "", &startFrom, &startTo, nil, nil)
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

		events, err := repo.FindEvent(ctx, db, "", nil, nil, &endFrom, &endTo)
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

		events, err := repo.FindEvent(ctx, db, "", nil, &targetTime, &targetTime, nil)
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
	crudRepo := NewEventCrudRepository(db)
	repo, err := NewEventRepository(crudRepo)
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
	crudRepo := NewEventCrudRepository(db)
	repo, err := NewEventRepository(crudRepo)
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

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
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

		events, err := repo.FindEvent(ctx, db, "", &from, &to, nil, nil)
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
