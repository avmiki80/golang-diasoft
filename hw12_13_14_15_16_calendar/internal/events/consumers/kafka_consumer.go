package consumers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/segmentio/kafka-go"
)

// KafkaConsumer простой consumer без retry логики
type KafkaConsumer struct {
	reader         *kafka.Reader
	handler        MessageHandler
	logger         logger.Logger
	groupID        string
	topic          string
	autoCommit     bool
	commitInterval time.Duration
}

// NewKafkaConsumer создает новый простой consumer без retry
func NewKafkaConsumer(
	cfg config.KafkaConf,
	topic string,
	handler MessageHandler,
	logg logger.Logger,
) *KafkaConsumer {
	// Определяем интервал коммита на основе конфигурации
	commitInterval := time.Duration(0) // По умолчанию ручной коммит
	autoCommit := false
	if cfg.Consumer.AutoCommit != nil && *cfg.Consumer.AutoCommit {
		commitInterval = 1 * time.Second // Автоматический коммит каждую секунду
		autoCommit = true
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.BootstrapServers,
		GroupID:        cfg.Consumer.GroupID,
		Topic:          topic,
		MinBytes:       10e3, // Можно вынести в конфиг
		MaxBytes:       10e6, // Можно вынести в конфиг
		StartOffset:    kafka.FirstOffset,
		CommitInterval: commitInterval,
	})

	return &KafkaConsumer{
		reader:         reader,
		handler:        handler,
		logger:         logg,
		groupID:        cfg.Consumer.GroupID,
		topic:          topic,
		autoCommit:     autoCommit,
		commitInterval: commitInterval,
	}
}

// Close закрывает все ресурсы
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// ProcessMessage обрабатывает сообщение без retry
func (c *KafkaConsumer) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	return c.handler.Handle(ctx, msg)
}

// Run запускает цикл обработки сообщений без retry
func (c *KafkaConsumer) Run(ctx context.Context) error {
	c.logger.Info(fmt.Sprintf("Starting Kafka consumer [%s]...", c.handler.GetName()))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutdown signal received, stopping consumer...")
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				c.logger.Error(fmt.Sprintf("Error reading message: %v", err))
				continue
			}

			c.logger.Info(fmt.Sprintf("Received message: topic=%s, partition=%d, offset=%d",
				msg.Topic, msg.Partition, msg.Offset))

			// Простая обработка без retry
			if err := c.ProcessMessage(ctx, msg); err != nil {
				c.logger.Error(fmt.Sprintf("Processing failed: %v", err))
				// Не коммитим оффсет, чтобы попробовать еще раз при следующем запуске
				continue
			}

			// Коммитим оффсет после успешной обработки
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to commit offset: %v", err))
			}
		}
	}
}
