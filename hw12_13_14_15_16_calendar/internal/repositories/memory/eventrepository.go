package memory

import (
	"context"
	"time"

	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/jmoiron/sqlx"
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
	return nil
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

func (r *EventRepository) FindEvent(_ context.Context, _ sqlx.ExtContext, userID string, startFrom, startTo, endFrom, endTo *time.Time) ([]events.Event, error) {
	r.crudRepo.mu.RLock()
	defer r.crudRepo.mu.RUnlock()

	result := make([]events.Event, 0, len(r.crudRepo.events))
	for _, event := range r.crudRepo.events {
		if userID != "" && event.UserID != userID {
			continue
		}

		if startFrom != nil && event.StartDate.Before(*startFrom) {
			continue
		}

		if startTo != nil && event.StartDate.After(*startTo) {
			continue
		}

		if endFrom != nil && event.EndDate.Before(*endFrom) {
			continue
		}

		if endTo != nil && event.EndDate.After(*endTo) {
			continue
		}

		result = append(result, event)
	}

	return result, nil
}
