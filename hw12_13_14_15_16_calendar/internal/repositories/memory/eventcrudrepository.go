package memory

import (
	"context"
	"sync"

	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EventCrudRepository struct {
	events map[string]events.Event
	mu     sync.RWMutex
}

func NewEventCrudRepository() *EventCrudRepository {
	return &EventCrudRepository{
		events: make(map[string]events.Event),
		mu:     sync.RWMutex{},
	}
}

func (r *EventCrudRepository) GetDB() *sqlx.DB {
	return nil // Memory storage doesn't have DB
}

func (r *EventCrudRepository) Create(_ context.Context, _ sqlx.ExtContext, event events.Event) (*events.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var err error
	newID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	event.ID = newID.String()
	if _, ok := r.events[event.ID]; ok {
		return nil, repositories.ErrEntityAlreadyExists
	}
	r.events[event.ID] = event
	return &event, nil
}

func (r *EventCrudRepository) Update(_ context.Context, _ sqlx.ExtContext, id string, event events.Event) (*events.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.events[id]; !ok {
		return nil, repositories.ErrEntityNotFound
	}
	r.events[id] = event
	return &event, nil
}

func (r *EventCrudRepository) Delete(_ context.Context, _ sqlx.ExtContext, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.events[id]; !ok {
		return repositories.ErrEntityNotFound
	}
	delete(r.events, id)
	return nil
}

func (r *EventCrudRepository) GetByID(_ context.Context, _ sqlx.ExtContext, id string) (*events.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if event, exists := r.events[id]; exists {
		return &event, nil
	}
	return nil, repositories.ErrEntityNotFound
}
