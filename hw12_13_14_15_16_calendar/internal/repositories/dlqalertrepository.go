package repositories

import (
	"context"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
)

type DLQAlertRepository interface {
	CrudRepository[domain.DLQAlert]
	GetDB() *sqlx.DB
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	GetUnresolvedCount(ctx context.Context, exec sqlx.ExtContext) (int, error)
	GetByStatus(ctx context.Context, exec sqlx.ExtContext, status string, limit int) ([]domain.DLQAlert, error)
}

type CompositeDLQAlertRepository interface {
	DLQAlertRepository
}
