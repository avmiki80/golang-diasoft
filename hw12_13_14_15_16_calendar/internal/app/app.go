package app

import (
	"context"
	"time"

	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
)

// Линтер так настоял

const (
	appName                = "app: "
	failedSendNotification = "app: failed to send notification: "
)

type Application interface {
	CreateEvent(ctx context.Context, event events.Event) (*events.Event, error)
	UpdateEvent(ctx context.Context, id string, event events.Event) (*events.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*events.Event, error)
	FindEvent(ctx context.Context, userID string, startFrom *time.Time, startTo *time.Time, endFrom *time.Time, endTo *time.Time) ([]events.Event, error)
}

type App struct {
	eventService  services.EventService
	notifyService services.NotificationService
	logger        logger.Logger
}

func New(eventService services.EventService, notifyService services.NotificationService, log logger.Logger) *App {
	return &App{
		eventService:  eventService,
		notifyService: notifyService,
		logger:        log,
	}
}

func (a *App) CreateEvent(ctx context.Context, event events.Event) (*events.Event, error) {
	a.logger.Debug(appName + "creating event " + event.ID)
	createdEvent, err := a.eventService.CreateEvent(ctx, event)
	if err != nil {
		a.logger.Error(appName + "failed to create event: " + err.Error())
		return nil, err
	}

	// Отправка уведомления (если сервис доступен)
	if a.notifyService != nil {
		if err := a.notifyService.NotifyEventCreated(ctx, event); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			a.logger.Warn(failedSendNotification + err.Error())
		}
	}

	a.logger.Info(appName + "event created successfully: " + event.ID)
	return createdEvent, nil
}

func (a *App) UpdateEvent(ctx context.Context, id string, event events.Event) (*events.Event, error) {
	a.logger.Debug(appName + "updating event " + id)
	updatedEvent, err := a.eventService.UpdateEvent(ctx, id, event)
	if err != nil {
		a.logger.Error(appName + "failed to update event: " + err.Error())
		return nil, err
	}

	// Отправка уведомления (если сервис доступен)
	if a.notifyService != nil {
		event.ID = id
		if err := a.notifyService.NotifyEventUpdated(ctx, event); err != nil {
			a.logger.Warn(failedSendNotification + err.Error())
		}
	}

	a.logger.Info(appName + "event updated successfully: " + id)
	return updatedEvent, nil
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Debug(appName + "deleting event " + id)

	if err := a.eventService.DeleteEvent(ctx, id); err != nil {
		a.logger.Error(appName + "failed to delete event: " + err.Error())
		return err
	}

	// Отправка уведомления (если сервис доступен)
	if a.notifyService != nil {
		if err := a.notifyService.NotifyEventDeleted(ctx, id); err != nil {
			a.logger.Warn(failedSendNotification + err.Error())
		}
	}

	a.logger.Info(appName + "event deleted successfully: " + id)
	return nil
}

func (a *App) GetEventByID(ctx context.Context, id string) (*events.Event, error) {
	a.logger.Debug(appName + "getting event " + id)
	return a.eventService.GetEventByID(ctx, id)
}

func (a *App) FindEvent(ctx context.Context, userID string, startFrom *time.Time, startTo *time.Time, endFrom *time.Time, endTo *time.Time) ([]events.Event, error) {
	a.logger.Debug(appName + "finding events")
	return a.eventService.FindEvent(ctx, userID, startFrom, startTo, endFrom, endTo)
}
