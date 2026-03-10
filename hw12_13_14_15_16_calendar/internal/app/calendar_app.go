package app

import (
	"context"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
)

// Линтер так настоял

const (
	appName = "app: "
)

type CalendarApplication interface {
	CreateEvent(ctx context.Context, event domain.Event) (*domain.Event, error)
	UpdateEvent(ctx context.Context, id string, event domain.Event) (*domain.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*domain.Event, error)
	FindEvent(ctx context.Context, userID string, startFrom, startTo, endFrom, endTo *time.Time, notificationSent *bool) ([]domain.Event, error)
}
type CalendarApp struct {
	eventService  services.EventService
	notifyService services.NotificationService
	logger        logger.Logger
}

func NewCalendarApp(eventService services.EventService, notifyService services.NotificationService, log logger.Logger) *CalendarApp {
	return &CalendarApp{
		eventService:  eventService,
		notifyService: notifyService,
		logger:        log,
	}
}

func (a *CalendarApp) CreateEvent(ctx context.Context, event domain.Event) (*domain.Event, error) {
	a.logger.Debug(appName + "creating event " + event.ID)
	createdEvent, err := a.eventService.CreateEvent(ctx, event)
	if err != nil {
		a.logger.Error(appName + "failed to create event: " + err.Error())
		return nil, err
	}
	a.logger.Info(appName + "event created successfully: " + event.ID)
	return createdEvent, nil
}

func (a *CalendarApp) UpdateEvent(ctx context.Context, id string, event domain.Event) (*domain.Event, error) {
	a.logger.Debug(appName + "updating event " + id)
	updatedEvent, err := a.eventService.UpdateEvent(ctx, id, event)
	if err != nil {
		a.logger.Error(appName + "failed to update event: " + err.Error())
		return nil, err
	}
	a.logger.Info(appName + "event updated successfully: " + id)
	return updatedEvent, nil
}

func (a *CalendarApp) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Debug(appName + "deleting event " + id)

	if err := a.eventService.DeleteEvent(ctx, id); err != nil {
		a.logger.Error(appName + "failed to delete event: " + err.Error())
		return err
	}
	a.logger.Info(appName + "event deleted successfully: " + id)
	return nil
}

func (a *CalendarApp) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	a.logger.Debug(appName + "getting event " + id)
	return a.eventService.GetEventByID(ctx, id)
}

func (a *CalendarApp) FindEvent(ctx context.Context, userID string, startFrom, startTo, endFrom, endTo *time.Time, notificationSent *bool) ([]domain.Event, error) {
	a.logger.Debug(appName + "finding events")
	return a.eventService.FindEvent(ctx, userID, startFrom, startTo, endFrom, endTo, notificationSent)
}
