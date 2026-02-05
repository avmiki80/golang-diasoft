package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type contextKey string

const txKey contextKey = "tx"

type TxManager interface {
	WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx *sqlx.Tx) error) error
	GetDB() *sqlx.DB
}

type txManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) TxManager {
	return &txManager{db: db}
}

func (tm *txManager) GetDB() *sqlx.DB {
	return tm.db
}

func (tm *txManager) WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx *sqlx.Tx) error) error {
	tx, err := tm.db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	ctx = context.WithValue(ctx, txKey, tx)

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction rollback error: %w, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
