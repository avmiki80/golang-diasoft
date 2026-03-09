package services

import (
	"context"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/jmoiron/sqlx"
)

type DLQAlertService interface {
	CreateAlert(ctx context.Context, alert domain.DLQAlert) (*domain.DLQAlert, error)
	GetUnresolvedCount(ctx context.Context) (int, error)
	GetByStatus(ctx context.Context, status string, limit int) ([]domain.DLQAlert, error)
	UpdateAlertStatus(ctx context.Context, id, status string) error
}

type dlqAlertService struct {
	repository repositories.CompositeDLQAlertRepository
	txManager  database.TxManager
}

func NewDLQAlertService(repo repositories.CompositeDLQAlertRepository, txManager database.TxManager) DLQAlertService {
	return &dlqAlertService{
		repository: repo,
		txManager:  txManager,
	}
}

func (s *dlqAlertService) CreateAlert(ctx context.Context, alert domain.DLQAlert) (*domain.DLQAlert, error) {
	var createdAlert *domain.DLQAlert
	err := s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		var err error
		createdAlert, err = s.repository.Create(ctx, exec, alert)
		return err
	})

	return createdAlert, err
}

func (s *dlqAlertService) GetUnresolvedCount(ctx context.Context) (int, error) {
	return s.repository.GetUnresolvedCount(ctx, s.getExecutor())
}

func (s *dlqAlertService) GetByStatus(ctx context.Context, status string, limit int) ([]domain.DLQAlert, error) {
	return s.repository.GetByStatus(ctx, s.getExecutor(), status, limit)
}

func (s *dlqAlertService) UpdateAlertStatus(ctx context.Context, id, status string) error {
	return s.executeWithTx(ctx, func(ctx context.Context, exec sqlx.ExtContext) error {
		alert, err := s.repository.GetByID(ctx, exec, id)
		if err != nil {
			return err
		}
		alert.Status = status
		_, err = s.repository.Update(ctx, exec, id, *alert)
		return err
	})
}

func (s *dlqAlertService) executeWithTx(ctx context.Context, fn func(context.Context, sqlx.ExtContext) error) error {
	if s.txManager == nil {
		return fn(ctx, nil)
	}
	return s.txManager.WithTransaction(ctx, nil, func(ctx context.Context, tx *sqlx.Tx) error {
		return fn(ctx, tx)
	})
}

func (s *dlqAlertService) getExecutor() sqlx.ExtContext {
	return s.repository.GetDB()
}
