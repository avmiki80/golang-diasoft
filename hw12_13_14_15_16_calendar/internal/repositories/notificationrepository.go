package repositories

import (
	"context"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository interface {
	FindNotification(ctx context.Context, exec sqlx.ExtContext, eventID, userID *string) ([]domain.Notification, error)
}

type CompositeNotificationRepository interface {
	CrudRepository[domain.Notification]
	NotificationRepository
	GetDB() *sqlx.DB
}
