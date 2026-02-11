//go:build integration
// +build integration

package db

import (
	"testing"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/testhelpers"
	"github.com/jmoiron/sqlx"
)

// SetupPostgresContainer создает и запускает PostgreSQL контейнер для тестов репозитория
func SetupPostgresContainer(t *testing.T) (*testhelpers.PostgresContainer, *sqlx.DB) {
	t.Helper()
	pc := testhelpers.SetupPostgresContainer(t, "calendar_repo_test")
	return pc, pc.DB
}

// cleanupTestData очищает тестовые данные из схемы public
func cleanupTestData(t *testing.T, db *sqlx.DB) {
	t.Helper()
	testhelpers.CleanupTestData(t, db)
}
