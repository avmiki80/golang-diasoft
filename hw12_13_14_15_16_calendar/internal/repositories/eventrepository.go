package repositories

import (
	"context"
	"time"

	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
)

type EventRepository interface {
	FindEvent(ctx context.Context, exec sqlx.ExtContext, userID string, startFrom, startTo, endFrom, endTo *time.Time) ([]events.Event, error)
}

type CompositeEventRepository interface {
	CrudRepository[events.Event]
	EventRepository
	GetDB() *sqlx.DB
}
