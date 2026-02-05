package repositories

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrEntityAlreadyExists = errors.New("entity already exists")
	ErrEntityNotFound      = errors.New("entity not found")
)

type CrudRepository[T any] interface {
	Create(ctx context.Context, exec sqlx.ExtContext, entity T) (*T, error)
	Update(ctx context.Context, exec sqlx.ExtContext, id string, entity T) (*T, error)
	Delete(ctx context.Context, exec sqlx.ExtContext, id string) error
	GetByID(ctx context.Context, exec sqlx.ExtContext, id string) (*T, error)
}
