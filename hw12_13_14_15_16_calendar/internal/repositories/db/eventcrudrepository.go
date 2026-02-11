package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/jmoiron/sqlx"
)

const (
	CreateQuery = `
        INSERT INTO events (title, description, start_date, end_date, user_id, offset_time)
        VALUES (:title, :description, :start_date, :end_date, :user_id, :offset_time)
        RETURNING id, created_at, updated_at
    `
	UpdateQuery = `
		UPDATE events 
		SET title = :title, 
		    description = :description, 
		    start_date = :start_date, 
		    end_date = :end_date, 
		    user_id = :user_id, 
		    offset_time = :offset_time,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	DeleteQuery  = "DELETE FROM events WHERE id = :id"
	GetByIDQuery = `
		SELECT id, title, description, start_date, end_date, user_id, offset_time, created_at, updated_at 
		FROM events 
		WHERE id = :id
	`
)

type EventCrudRepository struct {
	db *sqlx.DB
}

func NewEventCrudRepository(db *sqlx.DB) *EventCrudRepository {
	return &EventCrudRepository{db: db}
}

func (r *EventCrudRepository) GetDB() *sqlx.DB {
	return r.db
}

func (r *EventCrudRepository) Connect(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *EventCrudRepository) Close(_ context.Context) error {
	return r.db.Close()
}

func (r *EventCrudRepository) Create(ctx context.Context, exec sqlx.ExtContext, event events.Event) (*events.Event, error) {
	var createdEvent struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	query, args, err := sqlx.Named(CreateQuery, event)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.GetContext(ctx, exec, &createdEvent, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	event.ID = createdEvent.ID
	return &event, nil
}

func (r *EventCrudRepository) Update(ctx context.Context, exec sqlx.ExtContext, id string, event events.Event) (*events.Event, error) {
	event.ID = id

	query, args, err := sqlx.Named(UpdateQuery, event)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, repositories.ErrEntityNotFound
	}

	return &event, nil
}

func (r *EventCrudRepository) Delete(ctx context.Context, exec sqlx.ExtContext, id string) error {
	query, args, err := sqlx.Named(DeleteQuery, map[string]any{"id": id})
	if err != nil {
		return fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
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

func (r *EventCrudRepository) GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*events.Event, error) {
	var event events.Event

	query, args, err := sqlx.Named(GetByIDQuery, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.GetContext(ctx, exec, &event, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get event by id: %w", err)
	}

	return &event, nil
}
