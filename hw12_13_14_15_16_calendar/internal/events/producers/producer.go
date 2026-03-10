package producers

import (
	"context"
)

type Producer[T any] interface {
	SendMessage(ctx context.Context, message *T) error
	Close() error
}
