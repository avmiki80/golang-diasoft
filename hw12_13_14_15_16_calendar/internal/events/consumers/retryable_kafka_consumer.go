package consumers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/producers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/segmentio/kafka-go"
)

// RetryableKafkaConsumer consumer с поддержкой retry и DLQ
type RetryableKafkaConsumer struct {
	reader          *kafka.Reader
	dlqProducer     *producers.DLQProducer
	handler         MessageHandler
	logger          logger.Logger
	groupID         string
	topic           string
	autoCommit      bool
	commitInterval  time.Duration
	retryEnable     bool
	maxAttempts     int
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
	maxElapsedTime  time.Duration
}

// NewRetryableKafkaConsumer создает новый consumer с retry логикой
func NewRetryableKafkaConsumer(
	cfg config.KafkaConf,
	topic string,
	handler MessageHandler,
	dlqProducer *producers.DLQProducer,
	logg logger.Logger,
) *RetryableKafkaConsumer {
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

	return &RetryableKafkaConsumer{
		reader:          reader,
		dlqProducer:     dlqProducer,
		handler:         handler,
		logger:          logg,
		groupID:         cfg.Consumer.GroupID,
		topic:           topic,
		autoCommit:      autoCommit,
		commitInterval:  commitInterval,
		retryEnable:     cfg.Consumer.RetryEnable,
		maxAttempts:     cfg.Consumer.MaxAttempts,
		initialInterval: time.Duration(cfg.Consumer.InitialInterval) * time.Second,
		maxInterval:     time.Duration(cfg.Consumer.MaxInterval) * time.Second,
		multiplier:      cfg.Consumer.Multiplier,
		maxElapsedTime:  time.Duration(cfg.Consumer.MaxElapsedTime) * time.Second,
	}
}

// Close закрывает все ресурсы
func (c *RetryableKafkaConsumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	if c.dlqProducer != nil {
		return c.dlqProducer.Close()
	}
	return nil
}

// ProcessMessage обрабатывает сообщение с retry логикой
func (c *RetryableKafkaConsumer) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var lastErr error

	interval := c.initialInterval

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		c.logger.Info(fmt.Sprintf("[%s] Attempt %d/%d", c.handler.GetName(), attempt, c.maxAttempts))

		err := c.handler.Handle(ctx, msg)
		if err == nil {
			return nil // Успешная обработка
		}

		lastErr = err
		c.logger.Error(fmt.Sprintf("[%s] Failed attempt %d: %v", c.handler.GetName(), attempt, err))

		if attempt < c.maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
				// Экспоненциальная задержка
				interval = time.Duration(float64(interval) * c.multiplier)
				if interval > c.maxInterval {
					interval = c.maxInterval
				}
			}
		}
	}

	c.logger.Error(fmt.Sprintf("[%s] All %d attempts exhausted", c.handler.GetName(), c.maxAttempts))
	return lastErr
}

// Run запускает цикл обработки сообщений с retry и DLQ
func (c *RetryableKafkaConsumer) Run(ctx context.Context) error {
	c.logger.Info(fmt.Sprintf("Starting Retryable Kafka consumer [%s]...", c.handler.GetName()))

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

			var processingErr error
			var errorMessages []string

			// Обработка с retry (RetryableKafkaConsumer всегда использует retry)
			processingErr = c.ProcessMessage(ctx, msg)

			// Если обработка не удалась - отправляем в DLQ (если он настроен)
			if processingErr != nil {
				errorMessages = append(errorMessages, processingErr.Error())

				if c.dlqProducer != nil {
					c.logger.Error(fmt.Sprintf("Processing failed, sending to DLQ: %v", processingErr))

					if err := c.dlqProducer.SendToDLQ(ctx, msg.Value, msg, errorMessages); err != nil {
						c.logger.Error(fmt.Sprintf("Failed to send to DLQ: %v", err))
						// Не коммитим оффсет, чтобы попробовать еще раз
						continue
					}
				} else {
					c.logger.Error(fmt.Sprintf("Processing failed (DLQ not configured): %v", processingErr))
					// Не коммитим оффсет, чтобы попробовать еще раз
					continue
				}
			}

			// Коммитим оффсет после успешной обработки или отправки в DLQ
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to commit offset: %v", err))
			}
		}
	}
}
