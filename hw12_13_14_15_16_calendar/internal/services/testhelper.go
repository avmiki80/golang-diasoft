//go:build integration
// +build integration

package services

import (
	"testing"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/testhelpers"
	"github.com/jmoiron/sqlx"
)

// TestEnvironment содержит все необходимое для интеграционных тестов
type TestEnvironment struct {
	DB         *sqlx.DB
	TxManager  database.TxManager
	Repository repositories.CompositeEventRepository
	Service    EventService
}

// SetupTestEnvironment создает полное окружение для тестирования сервиса
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Используем общий helper для создания контейнера
	pc := testhelpers.SetupPostgresContainer(t, "calendar_service_test")

	// Создаем компоненты
	txManager := database.NewTxManager(pc.DB)
	crudRepo := db.NewEventCrudRepository(pc.DB)
	repository, err := db.NewEventRepository(crudRepo)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	service := NewEventService(repository, txManager)

	return &TestEnvironment{
		DB:         pc.DB,
		TxManager:  txManager,
		Repository: repository,
		Service:    service,
	}
}

// CleanupTestData очищает все данные из таблиц
func (env *TestEnvironment) CleanupTestData(t *testing.T) {
	t.Helper()
	testhelpers.CleanupTestData(t, env.DB)
}
