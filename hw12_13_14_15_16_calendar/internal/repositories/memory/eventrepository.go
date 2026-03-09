package memory

import (
	"context"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
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

func (r *EventRepository) FindEvent(_ context.Context, _ sqlx.ExtContext, userID string, startFrom, startTo, endFrom, endTo *time.Time, notificationSent *bool) ([]domain.Event, error) {
	r.crudRepo.mu.RLock()
	defer r.crudRepo.mu.RUnlock()

	result := make([]domain.Event, 0, len(r.crudRepo.events))
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
		if notificationSent != nil && event.NotificationSent != *notificationSent {
			continue
		}
		result = append(result, event)
	}

	return result, nil
}

func (r *EventRepository) NotificationSent(_ context.Context, _ sqlx.ExtContext, id string) error {
	r.crudRepo.mu.Lock()
	defer r.crudRepo.mu.Unlock()
	event, ok := r.crudRepo.events[id]
	if !ok {
		return repositories.ErrEntityNotFound
	}
	event.NotificationSent = true
	r.crudRepo.events[id] = event
	return nil
}

func (r *EventRepository) DeleteOldEvents(_ context.Context, _ sqlx.ExtContext, thresholdTime *time.Time) error {
	r.crudRepo.mu.Lock()
	defer r.crudRepo.mu.Unlock()

	idsToDelete := make([]string, 0)
	for _, event := range r.crudRepo.events {
		if event.EndDate.Before(*thresholdTime) || event.EndDate.Equal(*thresholdTime) {
			idsToDelete = append(idsToDelete, event.ID)
		}
	}

	for _, id := range idsToDelete {
		delete(r.crudRepo.events, id)
	}

	return nil
}

func (r *EventRepository) FindUpcomingEvent(_ context.Context, _ sqlx.ExtContext, userID string, thresholdTime *time.Time) ([]domain.Event, error) {
	r.crudRepo.mu.RLock()
	defer r.crudRepo.mu.RUnlock()

	result := make([]domain.Event, 0, len(r.crudRepo.events))
	for _, event := range r.crudRepo.events {
		if userID != "" && event.UserID != userID {
			continue
		}
		if event.NotificationSent {
			continue
		}
		timeNotificationEvent := event.StartDate.Add(-1 * event.OffsetTime)
		if thresholdTime != nil && timeNotificationEvent.After(*thresholdTime) {
			continue
		}

		result = append(result, event)
	}

	return result, nil
}
