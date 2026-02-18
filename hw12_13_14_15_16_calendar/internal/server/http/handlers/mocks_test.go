package handlers

import (
	"context"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockApplication - мок для app.Application
type MockApplication struct {
	mock.Mock
}

func (m *MockApplication) CreateEvent(ctx context.Context, event domain.Event) (*domain.Event, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockApplication) UpdateEvent(ctx context.Context, id string, event domain.Event) (*domain.Event, error) {
	args := m.Called(ctx, id, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockApplication) DeleteEvent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockApplication) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockApplication) FindEvent(ctx context.Context, userID string, startFrom, startTo, endFrom, endTo *time.Time) ([]domain.Event, error) {
	args := m.Called(ctx, userID, startFrom, startTo, endFrom, endTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Event), args.Error(1)
}

// MockLogger - мок для logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}
