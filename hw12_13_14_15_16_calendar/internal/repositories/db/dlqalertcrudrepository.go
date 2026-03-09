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
	CreateDLQAlertQuery = `
        INSERT INTO dlq_alerts (
            topic, original_topic, partition, message_offset, message_key, 
            error_message, error_count, original_message, failed_at, status
        )
        VALUES (
            :topic, :original_topic, :partition, :message_offset, :message_key,
            :error_message, :error_count, :original_message, :failed_at, :status
        )
        RETURNING id, created_at
    `
	UpdateDLQAlertQuery = `
		UPDATE dlq_alerts 
		SET status = :status
		WHERE id = :id
	`
	DeleteDLQAlertQuery = "DELETE FROM dlq_alerts WHERE id = :id"

	GetByIDDLQAlertQuery = `
		SELECT id, topic, original_topic, partition, message_offset, message_key,
		       error_message, error_count, original_message, failed_at, status, created_at
		FROM dlq_alerts 
		WHERE id = :id
	`
	GetUnresolvedCountQuery = `
		SELECT COUNT(*) FROM dlq_alerts WHERE status = 'new'
	`
	GetByStatusQuery = `
		SELECT id, topic, original_topic, partition, message_offset, message_key,
		       error_message, error_count, original_message, failed_at, status, created_at
		FROM dlq_alerts 
		WHERE status = :status
		ORDER BY created_at DESC
		LIMIT :limit
	`
)

type DLQAlertCrudRepository struct {
	db *sqlx.DB
}

func NewDLQAlertCrudRepository(db *sqlx.DB) *DLQAlertCrudRepository {
	return &DLQAlertCrudRepository{db: db}
}

func (r *DLQAlertCrudRepository) GetDB() *sqlx.DB {
	return r.db
}

func (r *DLQAlertCrudRepository) Connect(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *DLQAlertCrudRepository) Close(_ context.Context) error {
	return r.db.Close()
}

func (r *DLQAlertCrudRepository) Create(ctx context.Context, exec sqlx.ExtContext, alert domain.DLQAlert) (*domain.DLQAlert, error) {
	var created struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
	}

	query, args, err := sqlx.Named(CreateDLQAlertQuery, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.GetContext(ctx, exec, &created, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create dlq alert: %w", err)
	}
	alert.ID = created.ID
	return &alert, nil
}

func (r *DLQAlertCrudRepository) Update(ctx context.Context, exec sqlx.ExtContext, id string, alert domain.DLQAlert) (*domain.DLQAlert, error) {
	alert.ID = id

	query, args, err := sqlx.Named(UpdateDLQAlertQuery, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update dlq alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, repositories.ErrEntityNotFound
	}

	return &alert, nil
}

func (r *DLQAlertCrudRepository) Delete(ctx context.Context, exec sqlx.ExtContext, id string) error {
	query, args, err := sqlx.Named(DeleteDLQAlertQuery, map[string]any{"id": id})
	if err != nil {
		return fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete dlq alert: %w", err)
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

func (r *DLQAlertCrudRepository) GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*domain.DLQAlert, error) {
	var alert domain.DLQAlert

	query, args, err := sqlx.Named(GetByIDDLQAlertQuery, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.GetContext(ctx, exec, &alert, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get dlq alert by id: %w", err)
	}

	return &alert, nil
}

func (r *DLQAlertCrudRepository) GetUnresolvedCount(ctx context.Context, exec sqlx.ExtContext) (int, error) {
	var count int
	err := exec.QueryRowxContext(ctx, GetUnresolvedCountQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unresolved count: %w", err)
	}
	return count, nil
}

func (r *DLQAlertCrudRepository) GetByStatus(ctx context.Context, exec sqlx.ExtContext, status string, limit int) ([]domain.DLQAlert, error) {
	var alerts []domain.DLQAlert

	query, args, err := sqlx.Named(GetByStatusQuery, map[string]any{
		"status": status,
		"limit":  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	query = r.db.Rebind(query)

	err = sqlx.SelectContext(ctx, exec, &alerts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts by status: %w", err)
	}

	return alerts, nil
}
