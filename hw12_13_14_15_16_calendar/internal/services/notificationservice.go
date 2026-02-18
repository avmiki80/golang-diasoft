package services

import (
	"context"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
)

type NotificationService interface {
	// NotifyEventCreated отправляет уведомление о создании события.
	NotifyEventCreated(ctx context.Context, event domain.Event) error

	// NotifyEventUpdated отправляет уведомление об обновлении события.
	NotifyEventUpdated(ctx context.Context, event domain.Event) error

	// NotifyEventDeleted отправляет уведомление об удалении события.
	NotifyEventDeleted(ctx context.Context, eventID string) error
}
