package database

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // PostgreSQL driver
	"github.com/jmoiron/sqlx"
)

type ConnectionConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func DefaultConnectionConfig(dsn string) ConnectionConfig {
	return ConnectionConfig{
		DSN:             dsn,
		MaxOpenConns:    15,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func NewConnection(config ConnectionConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
