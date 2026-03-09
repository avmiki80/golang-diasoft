package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/jmoiron/sqlx"
)

const (
	FindEventsQueryBase = `
		SELECT id, title, description, start_date, end_date, user_id, offset_time, notification_sent, created_at, updated_at 
		FROM events
	`
	NotificationSentQuery = `
		UPDATE events 
		SET notification_sent = true 
		WHERE id = :id
	`
	DeleteOldEventsQuery = `DELETE FROM events WHERE end_date <= :thresholdTime`
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

func (r *EventRepository) Create(ctx context.Context, exec sqlx.ExtContext, event domain.Event) (*domain.Event, error) {
	return r.crudRepo.Create(ctx, exec, event)
}

func (r *EventRepository) Update(ctx context.Context, exec sqlx.ExtContext, id string, event domain.Event) (*domain.Event, error) {
	return r.crudRepo.Update(ctx, exec, id, event)
}

func (r *EventRepository) Delete(ctx context.Context, exec sqlx.ExtContext, id string) error {
	return r.crudRepo.Delete(ctx, exec, id)
}

func (r *EventRepository) GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*domain.Event, error) {
	return r.crudRepo.GetByID(ctx, exec, id)
}

func (r *EventRepository) FindEvent(ctx context.Context, exec sqlx.ExtContext, userID string, startFrom, startTo, endFrom, endTo *time.Time, notificationSent *bool) ([]domain.Event, error) {
	var eventsList []domain.Event

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

	if notificationSent != nil {
		whereClauses = append(whereClauses, "notification_sent = :notificationSent")
		params["notificationSent"] = *notificationSent
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

func (r *EventRepository) NotificationSent(ctx context.Context, exec sqlx.ExtContext, id string) error {
	query, args, err := sqlx.Named(NotificationSentQuery, map[string]any{"id": id})
	if err != nil {
		return fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.crudRepo.GetDB().Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repositories.ErrEntityNotFound
	}

	return nil
}

func (r *EventRepository) DeleteOldEvents(ctx context.Context, exec sqlx.ExtContext, thresholdTime *time.Time) error {
	query, args, err := sqlx.Named(DeleteOldEventsQuery, map[string]any{"thresholdTime": thresholdTime})
	if err != nil {
		return fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.crudRepo.GetDB().Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	return nil
}

func (r *EventRepository) FindUpcomingEvent(ctx context.Context, exec sqlx.ExtContext, userID string, thresholdTime *time.Time) ([]domain.Event, error) {
	var eventsList []domain.Event

	whereClauses := []string{"1=1"}
	params := make(map[string]any)

	if userID != "" {
		whereClauses = append(whereClauses, "user_id = :userID")
		params["userID"] = userID
	}

	if thresholdTime != nil {
		whereClauses = append(whereClauses, "start_date - make_interval(secs => offset_time / 1000000000.0) <= :thresholdTime")
		params["thresholdTime"] = *thresholdTime
	}

	whereClauses = append(whereClauses, "notification_sent = :notificationSent")
	params["notificationSent"] = false

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
		return nil, fmt.Errorf("failed to find upcoming events: %w", err)
	}

	return eventsList, nil
}
