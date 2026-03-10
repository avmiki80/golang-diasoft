package services

import (
	"context"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/jmoiron/sqlx"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, notification domain.Notification) (*domain.Notification, error)
	UpdateNotification(ctx context.Context, id string, notification domain.Notification) (*domain.Notification, error)
	DeleteNotification(ctx context.Context, id string) error
	GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error)
}

type notificationService struct {
	repository repositories.CompositeNotificationRepository
	txManager  database.TxManager
}

func NewNotificationService(repo repositories.CompositeNotificationRepository, txManager database.TxManager) NotificationService {
	return &notificationService{
		repository: repo,
		txManager:  txManager,
	}
}

func (s *notificationService) CreateNotification(ctx context.Context, notification domain.Notification) (*domain.Notification, error) {
	if err := s.validateNotification(notification); err != nil {
		return nil, err
	}

	var createdNotification *domain.Notification
	err := s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		// Проверка идемпотентности: ищем существующее уведомление по event_id
		existingNotifications, err := s.repository.FindNotification(ctx, exec, &notification.EventID, nil)
		if err != nil {
			return err
		}

		// Если уведомление уже существует, возвращаем его
		if len(existingNotifications) > 0 {
			createdNotification = &existingNotifications[0]
			return nil
		}

		// Создаем новое уведомление
		createdNotification, err = s.repository.Create(ctx, exec, notification)
		return err
	})

	return createdNotification, err
}

func (s *notificationService) UpdateNotification(ctx context.Context, id string, notification domain.Notification) (*domain.Notification, error) {
	if id == "" {
		return nil, ErrInvalidEventID
	}

	if err := s.validateNotification(notification); err != nil {
		return nil, err
	}

	var updatedNotification *domain.Notification
	err := s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		_, err := s.repository.GetByID(ctx, exec, id)
		if err != nil {
			if err.Error() == EntityNotFound {
				return ErrEventNotFound
			}
			return err
		}

		updatedNotification, err = s.repository.Update(ctx, exec, id, notification)
		return err
	})

	return updatedNotification, err
}

func (s *notificationService) DeleteNotification(ctx context.Context, id string) error {
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

func (s *notificationService) GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error) {
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

func (s *notificationService) executeWithTx(ctx context.Context, fn func(context.Context, sqlx.ExtContext) error) error {
	if s.txManager == nil {
		return fn(ctx, nil)
	}
	return s.txManager.WithTransaction(ctx, nil, func(ctx context.Context, tx *sqlx.Tx) error {
		return fn(ctx, tx)
	})
}

func (s *notificationService) getExecutor() sqlx.ExtContext {
	return s.repository.GetDB()
}

func (s *notificationService) validateNotification(notification domain.Notification) error {
	if notification.Title == "" {
		return ErrInvalidEventTitle
	}

	if notification.UserID == "" {
		return ErrInvalidUserID
	}

	if notification.EventID == "" {
		return ErrInvalidEventID
	}

	if notification.StartDate.IsZero() {
		return ErrInvalidStartDate
	}

	if notification.EndDate.IsZero() {
		return ErrInvalidEndDate
	}

	if !notification.EndDate.After(notification.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}
