package consumers

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Close() error
	Run(ctx context.Context) error
	ProcessMessage(ctx context.Context, msg kafka.Message) error
}

// MessageHandler интерфейс для обработки конкретных типов сообщений
type MessageHandler interface {
	// Handle обрабатывает десериализованное сообщение
	Handle(ctx context.Context, msg kafka.Message) error
	// GetName возвращает имя обработчика для логирования
	GetName() string
}
