//go:build integration
// +build integration

package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer управляет PostgreSQL контейнером для тестов
type PostgresContainer struct {
	container *postgres.PostgresContainer
	connStr   string
}

// SetupPostgresContainer создает и запускает PostgreSQL контейнер
func SetupPostgresContainer(t *testing.T) (*PostgresContainer, *sqlx.DB) {
	t.Helper()

	ctx := context.Background()

	// Создаем PostgreSQL контейнер
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("calendar"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Получаем connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Подключаемся к БД
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Применяем миграции из файлов
	if err := applyMigrationsFromFiles(db); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	pc := &PostgresContainer{
		container: pgContainer,
		connStr:   connStr,
	}

	// Регистрируем cleanup
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	return pc, db
}

// applyMigrationsFromFiles читает и применяет миграции из папки migrations
func applyMigrationsFromFiles(db *sqlx.DB) error {
	// Получаем путь к корню проекта (3 уровня вверх от internal/repositories/db)
	projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	migrationsDir := filepath.Join(projectRoot, "migrations")

	// Читаем файлы миграций в порядке
	migrationFiles := []string{
		"00001_create_events_table.sql",
		"00002_add_timestamps_to_events.sql",
	}

	for _, filename := range migrationFiles {
		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Извлекаем SQL из goose формата (между -- +goose Up и -- +goose Down)
		sql := extractUpMigration(string(content))
		if sql == "" {
			return fmt.Errorf("no SQL found in migration file %s", filename)
		}

		// Применяем миграцию к схеме public
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}
	}

	return nil
}

// extractUpMigration извлекает SQL для миграции "Up" из goose формата
func extractUpMigration(content string) string {
	lines := strings.Split(content, "\n")
	var sqlLines []string
	inUpSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "-- +goose Up") {
			inUpSection = true
			continue
		}

		if strings.Contains(trimmed, "-- +goose Down") {
			break
		}

		if inUpSection && !strings.Contains(trimmed, "-- +goose StatementBegin") &&
			!strings.Contains(trimmed, "-- +goose StatementEnd") {
			sqlLines = append(sqlLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(sqlLines, "\n"))
}

// cleanupTestData очищает тестовые данные из схемы public
func cleanupTestData(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE TABLE public.events CASCADE")
	if err != nil {
		t.Fatalf("Failed to cleanup test data: %v", err)
	}
}
