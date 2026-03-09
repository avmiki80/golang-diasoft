package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
)

const (
	FindNotificationsQueryBase = `
		SELECT id, title, start_date, end_date, user_id, event_id, created_at 
		FROM notifications
	`
)

type NotificationRepository struct {
	crudRepo *NotificationCrudRepository
}

func NewNotificationRepository(crudRepo *NotificationCrudRepository) (*NotificationRepository, error) {
	return &NotificationRepository{
		crudRepo: crudRepo,
	}, nil
}

func (r *NotificationRepository) GetDB() *sqlx.DB {
	return r.crudRepo.GetDB()
}

func (r *NotificationRepository) Create(ctx context.Context, exec sqlx.ExtContext, notification domain.Notification) (*domain.Notification, error) {
	return r.crudRepo.Create(ctx, exec, notification)
}

func (r *NotificationRepository) Update(ctx context.Context, exec sqlx.ExtContext, id string, notification domain.Notification) (*domain.Notification, error) {
	return r.crudRepo.Update(ctx, exec, id, notification)
}

func (r *NotificationRepository) Delete(ctx context.Context, exec sqlx.ExtContext, id string) error {
	return r.crudRepo.Delete(ctx, exec, id)
}

func (r *NotificationRepository) GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*domain.Notification, error) {
	return r.crudRepo.GetByID(ctx, exec, id)
}

func (r *NotificationRepository) FindNotification(ctx context.Context, exec sqlx.ExtContext, eventID, userID *string) ([]domain.Notification, error) {
	var notificationsList []domain.Notification

	whereClauses := []string{"1=1"}
	params := make(map[string]any)

	if eventID != nil && *eventID != "" {
		whereClauses = append(whereClauses, "event_id = :eventID")
		params["eventID"] = *eventID
	}

	if userID != nil && *userID != "" {
		whereClauses = append(whereClauses, "user_id = :userID")
		params["userID"] = *userID
	}

	query := fmt.Sprintf("%s WHERE %s ORDER BY created_at",
		FindNotificationsQueryBase,
		strings.Join(whereClauses, " AND "))

	namedQuery, args, err := sqlx.Named(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}

	namedQuery = r.crudRepo.GetDB().Rebind(namedQuery)

	err = sqlx.SelectContext(ctx, exec, &notificationsList, namedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find notifications: %w", err)
	}

	return notificationsList, nil
}
