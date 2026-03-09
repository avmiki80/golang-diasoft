package mapper

import (
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/messages"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/producers"
)

func NotificationMessageToDomain(msg messages.Message[producers.NotificationMessage]) domain.Notification {
	return domain.Notification{
		Title:     msg.Payload.Title,
		StartDate: msg.Payload.StartDate,
		EndDate:   msg.Payload.EndDate,
		EventID:   msg.Payload.EventID,
		UserID:    msg.Payload.UserID,
	}
}

func DomainNotificationToMessage(notification domain.Notification) producers.SentNotificationMessage {
	return producers.SentNotificationMessage{
		EventID: notification.EventID,
		UserID:  notification.UserID,
	}
}
