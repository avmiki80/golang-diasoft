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
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/segmentio/kafka-go"
)

type SentNotificationHandler struct {
	eventService services.EventService
	logger       logger.Logger
}

func NewSentNotificationHandler(eventService services.EventService, logg logger.Logger) *SentNotificationHandler {
	return &SentNotificationHandler{
		eventService: eventService,
		logger:       logg,
	}
}

// Handle обрабатывает сообщение с уведомлением
func (h *SentNotificationHandler) Handle(ctx context.Context, msg kafka.Message) error {
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

	var sentNotification messages.Message[producers.SentNotificationMessage]

	if err := json.Unmarshal(msg.Value, &sentNotification); err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	h.logger.Info(fmt.Sprintf("Processing notification: ID=%s, EventID=%s",
		msgID,
		sentNotification.Payload.EventID))

	if command == "sent-notification" {
		err := h.eventService.NotificationSent(ctx, sentNotification.Payload.EventID)
		if err != nil {
			return err
		}
	}

	// Симуляция обработки
	time.Sleep(100 * time.Millisecond)

	h.logger.Info(fmt.Sprintf("Notification sent successfully: ID=%s", msgID))
	return nil
}

func (h *SentNotificationHandler) GetName() string {
	return "SentNotificationHandler"
}
