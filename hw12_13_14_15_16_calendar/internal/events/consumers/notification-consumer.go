package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/messages"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/producers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/mapper"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/segmentio/kafka-go"
)

// NotificationHandler обрабатывает уведомления о событиях
type NotificationHandler struct {
	notificationService      services.NotificationService
	sentNotificationProducer *producers.SentNotificationProducer
	logger                   logger.Logger
}

func NewNotificationHandler(
	notificationService services.NotificationService,
	sentNotificationProducer *producers.SentNotificationProducer,
	logg logger.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService:      notificationService,
		sentNotificationProducer: sentNotificationProducer,
		logger:                   logg,
	}
}

// Handle обрабатывает сообщение с уведомлением
func (h *NotificationHandler) Handle(ctx context.Context, msg kafka.Message) error {
	if msg.Headers == nil {
		return errors.New("headers not found")
	}

	msgID, ok := events.GetHeaderValue(msg.Headers, "message-id")

	if !ok {
		return errors.New("message-id not found")
	}

	command, ok := events.GetHeaderValue(msg.Headers, "command")
	if !ok {
		return errors.New("command not found")
	}

	var notification messages.Message[producers.NotificationMessage]

	if err := json.Unmarshal(msg.Value, &notification); err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	h.logger.Info(fmt.Sprintf("Processing notification: ID=%s, EventID=%s, UserID=%s, Title=%s",
		msgID,
		notification.Payload.EventID,
		notification.Payload.UserID,
		notification.Payload.Title))

	if command == "notification" {
		if notification.Payload.Title == "тест retry" {
			return errors.New("тестирую retry")
		}
		createdNotification, err := h.notificationService.CreateNotification(ctx, mapper.NotificationMessageToDomain(notification))
		if err != nil {
			return err
		}
		msg := mapper.DomainNotificationToMessage(*createdNotification)
		err = h.sentNotificationProducer.SendMessage(ctx, &msg)
		if err != nil {
			return err
		}
	}

	// Симуляция обработки
	time.Sleep(100 * time.Millisecond)

	h.logger.Info(fmt.Sprintf("Notification sent successfully: ID=%s", msgID))
	return nil
}

// GetName возвращает имя обработчика
func (h *NotificationHandler) GetName() string {
	return "NotificationHandler"
}
