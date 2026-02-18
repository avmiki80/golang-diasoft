package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/jmoiron/sqlx"
)

const EntityNotFound = "entity not found"

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrInvalidEventID    = errors.New("event ID cannot be empty")
	ErrInvalidEventTitle = errors.New("event title cannot be empty")
	ErrDateBusy          = errors.New("date is busy")
	ErrInvalidUserID     = errors.New("user ID cannot be empty")
	ErrInvalidStartDate  = errors.New("start date cannot be empty")
	ErrInvalidEndDate    = errors.New("end date cannot be empty")
	ErrInvalidDateRange  = errors.New("end date must be after start date")
)

type EventService interface {
	CreateEvent(ctx context.Context, event events.Event) (*events.Event, error)
	UpdateEvent(ctx context.Context, id string, event events.Event) (*events.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*events.Event, error)
	FindEvent(ctx context.Context, userID string, startFrom, startTo, endFrom, endTo *time.Time) ([]events.Event, error)
}

type eventService struct {
	repository repositories.CompositeEventRepository
	txManager  database.TxManager
}

func NewEventService(repo repositories.CompositeEventRepository, txManager database.TxManager) EventService {
	return &eventService{
		repository: repo,
		txManager:  txManager,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, event events.Event) (*events.Event, error) {
	if err := s.validateEvent(event); err != nil {
		return nil, err
	}

	var createdEvent *events.Event
	err := s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		if err := s.checkCrossEvents(ctx, exec, event); err != nil {
			return err
		}
		var err error
		createdEvent, err = s.repository.Create(ctx, exec, event)
		return err
	})

	return createdEvent, err
}

func (s *eventService) UpdateEvent(ctx context.Context, id string, event events.Event) (*events.Event, error) {
	if id == "" {
		return nil, ErrInvalidEventID
	}

	if err := s.validateEvent(event); err != nil {
		return nil, err
	}

	var updatedEvent *events.Event
	err := s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		_, err := s.repository.GetByID(ctx, exec, id)
		if err != nil {
			if err.Error() == EntityNotFound {
				return ErrEventNotFound
			}
			return err
		}
		if err := s.checkCrossEvents(ctx, exec, event); err != nil {
			return err
		}

		updatedEvent, err = s.repository.Update(ctx, exec, id, event)
		return err
	})

	return updatedEvent, err
}

func (s *eventService) DeleteEvent(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidEventID
	}

	return s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		err := s.repository.Delete(ctx, exec, id)
		if err != nil {
			if err.Error() == EntityNotFound {
				return ErrEventNotFound
			}
			return err
		}
		return err
	})
}

func (s *eventService) GetEventByID(ctx context.Context, id string) (*events.Event, error) {
	if id == "" {
		return nil, ErrInvalidEventID
	}
	founded, err := s.repository.GetByID(ctx, s.getExecutor(), id)
	if err != nil {
		if err.Error() == EntityNotFound {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return founded, nil
}

func (s *eventService) FindEvent(ctx context.Context, userID string, startFrom, startTo, endFrom, endTo *time.Time) ([]events.Event, error) {
	return s.repository.FindEvent(ctx, s.getExecutor(), userID, startFrom, startTo, endFrom, endTo)
}

func (s *eventService) checkCrossEvents(ctx context.Context, exec sqlx.ExtContext, event events.Event) error {
	startTo := event.EndDate.Add(-time.Nanosecond)
	endFrom := event.StartDate.Add(time.Nanosecond)

	crossEvents, err := s.repository.FindEvent(ctx, exec, event.UserID, nil, &startTo, &endFrom, nil)
	if err != nil {
		return fmt.Errorf("failed to check cross events: %w", err)
	}

	for _, e := range crossEvents {
		if e.ID != event.ID {
			return ErrDateBusy
		}
	}

	return nil
}

func (s *eventService) validateEvent(event events.Event) error {
	if event.Title == "" {
		return ErrInvalidEventTitle
	}

	if event.UserID == "" {
		return ErrInvalidUserID
	}

	if event.StartDate.IsZero() {
		return ErrInvalidStartDate
	}

	if event.EndDate.IsZero() {
		return ErrInvalidEndDate
	}

	if !event.EndDate.After(event.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}

func (s *eventService) executeWithTx(ctx context.Context, fn func(context.Context, sqlx.ExtContext) error) error {
	if s.txManager == nil {
		return fn(ctx, nil)
	}
	return s.txManager.WithTransaction(ctx, nil, func(ctx context.Context, tx *sqlx.Tx) error {
		return fn(ctx, tx)
	})
}

func (s *eventService) getExecutor() sqlx.ExtContext {
	return s.repository.GetDB()
}
