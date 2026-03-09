package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/jmoiron/sqlx"
)

const (
	CreateNotificationQuery = `
        INSERT INTO notifications (title, start_date, end_date, user_id, event_id)
        VALUES (:title, :start_date, :end_date, :user_id, :event_id)
        RETURNING id, created_at
    `
	UpdateNotificationQuery = `
		UPDATE notifications 
		SET title = :title, 
		    start_date = :start_date, 
		    end_date = :end_date, 
		    user_id = :user_id, 
		    event_id = :event_id
		WHERE id = :id
	`
	DeleteNotificationQuery  = "DELETE FROM notifications WHERE id = :id"
	GetByIDNotificationQuery = `
		SELECT id, title, start_date, end_date, user_id, event_id, created_at 
		FROM notifications 
		WHERE id = :id
	`
)

type NotificationCrudRepository struct {
	db *sqlx.DB
}

func NewNotificationCrudRepository(db *sqlx.DB) *NotificationCrudRepository {
	return &NotificationCrudRepository{db: db}
}

func (r *NotificationCrudRepository) GetDB() *sqlx.DB {
	return r.db
}

func (r *NotificationCrudRepository) Connect(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *NotificationCrudRepository) Close(_ context.Context) error {
	return r.db.Close()
}

func (r *NotificationCrudRepository) Create(ctx context.Context, exec sqlx.ExtContext, notification domain.Notification) (*domain.Notification, error) {
	var createdNotification struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
	}

	query, args, err := sqlx.Named(CreateNotificationQuery, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.GetContext(ctx, exec, &createdNotification, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	notification.ID = createdNotification.ID
	return &notification, nil
}

func (r *NotificationCrudRepository) Update(ctx context.Context, exec sqlx.ExtContext, id string, notification domain.Notification) (*domain.Notification, error) {
	notification.ID = id

	query, args, err := sqlx.Named(UpdateNotificationQuery, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, repositories.ErrEntityNotFound
	}

	return &notification, nil
}

func (r *NotificationCrudRepository) Delete(ctx context.Context, exec sqlx.ExtContext, id string) error {
	query, args, err := sqlx.Named(DeleteNotificationQuery, map[string]any{"id": id})
	if err != nil {
		return fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
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

func (r *NotificationCrudRepository) GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*domain.Notification, error) {
	var notification domain.Notification

	query, args, err := sqlx.Named(GetByIDNotificationQuery, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.GetContext(ctx, exec, &notification, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get notification by id: %w", err)
	}

	return &notification, nil
}
