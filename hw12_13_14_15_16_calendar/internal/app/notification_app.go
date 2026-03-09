package app

import (
	"context"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/producers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/mapper"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
)

type NotificationApplication interface {
	DeleteOldEvents(ctx context.Context) error
	SendNotification(ctx context.Context) error
	Close() error
}

type NotificationApp struct {
	eventService         services.EventService
	notificationProducer *producers.NotificationProducer
	logger               logger.Logger
}

func NewNotificationApp(service services.EventService, producer *producers.NotificationProducer, log logger.Logger) *NotificationApp {
	return &NotificationApp{
		eventService:         service,
		notificationProducer: producer,
		logger:               log,
	}
}

func (a *NotificationApp) DeleteOldEvents(ctx context.Context) error {
	a.logger.Debug("deleting old events")
	oneYearOld := time.Now().AddDate(-1, 0, 0)
	if err := a.eventService.DeleteOldEvents(ctx, &oneYearOld); err != nil {
		a.logger.Error("failed to delete old events: " + err.Error())
		return err
	}
	a.logger.Info("old events deleted successfully")
	return nil
}

func (a *NotificationApp) SendNotification(ctx context.Context) error {
	a.logger.Debug("sending notifications")
	currentTime := time.Now()

	events, err := a.eventService.FindUpcomingEvent(ctx, "", &currentTime)
	if err != nil {
		a.logger.Error("failed to find upcoming events: " + err.Error())
		return err
	}

	if len(events) == 0 {
		a.logger.Debug("no upcoming events found")
		return nil
	}

	a.logger.Info("found " + string(rune(len(events))) + " upcoming events")

	for _, event := range events {
		a.logger.Debug("sending notification for event: " + event.Title)
		msg, err := mapper.DomainToNotificationMessage(event)
		if err != nil {
			a.logger.Error("failed to map event to notification: " + err.Error())
			continue
		}

		if err := a.notificationProducer.SendMessage(ctx, &msg); err != nil {
			a.logger.Error("failed to send notification for event " + event.ID + ": " + err.Error())
			continue
		}

		// Отметить что уведомление отправлено
		//if err := a.eventService.NotificationSent(ctx, event.ID); err != nil {
		//	a.logger.Error("failed to mark notification as sent for event " + event.ID + ": " + err.Error())
		//}
	}

	a.logger.Info("notifications sent successfully")
	return nil
}

func (a *NotificationApp) Close() error {
	return a.notificationProducer.Close()
}
