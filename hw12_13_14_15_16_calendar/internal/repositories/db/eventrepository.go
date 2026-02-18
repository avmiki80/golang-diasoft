package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
)

const (
	FindEventsQueryBase = `
		SELECT id, title, description, start_date, end_date, user_id, offset_time, created_at, updated_at 
		FROM events
	`
)

type EventRepository struct {
	crudRepo *EventCrudRepository
}

func NewEventRepository(crudRepo *EventCrudRepository) (*EventRepository, error) {
	return &EventRepository{
		crudRepo: crudRepo,
	}, nil
}

func (r *EventRepository) GetDB() *sqlx.DB {
	return r.crudRepo.GetDB()
}

func (r *EventRepository) Create(ctx context.Context, exec sqlx.ExtContext, event events.Event) (*events.Event, error) {
	return r.crudRepo.Create(ctx, exec, event)
}

func (r *EventRepository) Update(ctx context.Context, exec sqlx.ExtContext, id string, event events.Event) (*events.Event, error) {
	return r.crudRepo.Update(ctx, exec, id, event)
}

func (r *EventRepository) Delete(ctx context.Context, exec sqlx.ExtContext, id string) error {
	return r.crudRepo.Delete(ctx, exec, id)
}

func (r *EventRepository) GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*events.Event, error) {
	return r.crudRepo.GetByID(ctx, exec, id)
}

func (r *EventRepository) FindEvent(ctx context.Context, exec sqlx.ExtContext, userID string, startFrom, startTo, endFrom, endTo *time.Time) ([]events.Event, error) {
	var eventsList []events.Event

	whereClauses := []string{"1=1"}
	params := make(map[string]any)

	if userID != "" {
		whereClauses = append(whereClauses, "user_id = :userID")
		params["userID"] = userID
	}

	if startFrom != nil {
		whereClauses = append(whereClauses, "start_date >= :startFrom")
		params["startFrom"] = *startFrom
	}

	if startTo != nil {
		whereClauses = append(whereClauses, "start_date <= :startTo")
		params["startTo"] = *startTo
	}

	if endFrom != nil {
		whereClauses = append(whereClauses, "end_date >= :endFrom")
		params["endFrom"] = *endFrom
	}

	if endTo != nil {
		whereClauses = append(whereClauses, "end_date <= :endTo")
		params["endTo"] = *endTo
	}

	query := fmt.Sprintf("%s WHERE %s ORDER BY start_date",
		FindEventsQueryBase,
		strings.Join(whereClauses, " AND "))

	namedQuery, args, err := sqlx.Named(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	namedQuery = r.crudRepo.GetDB().Rebind(namedQuery)

	err = sqlx.SelectContext(ctx, exec, &eventsList, namedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}

	return eventsList, nil
}
