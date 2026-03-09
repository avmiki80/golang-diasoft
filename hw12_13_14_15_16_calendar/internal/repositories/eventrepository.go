package repositories

import (
	"context"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
)

type EventRepository interface {
	FindEvent(ctx context.Context, exec sqlx.ExtContext, userID string, startFrom, startTo, endFrom, endTo *time.Time, notificationSent *bool) ([]domain.Event, error)
	FindUpcomingEvent(ctx context.Context, exec sqlx.ExtContext, userID string, thresholdTime *time.Time) ([]domain.Event, error)
	NotificationSent(ctx context.Context, exec sqlx.ExtContext, id string) error
	DeleteOldEvents(ctx context.Context, exec sqlx.ExtContext, thresholdTime *time.Time) error
}

type CompositeEventRepository interface {
	CrudRepository[domain.Event]
	EventRepository
	GetDB() *sqlx.DB
}
