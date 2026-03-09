package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/messages"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/segmentio/kafka-go"
)

// DLQProducer управляет отправкой сообщений в Dead Letter Queue
type DLQProducer struct {
	writer *kafka.Writer
	logger logger.Logger
}

// NewDLQProducer создает новый producer для DLQ
func NewDLQProducer(brokers []string, topic string, logg logger.Logger) *DLQProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic + "-dlq",
		Balancer: &kafka.LeastBytes{},
		Async:    false,
	}

	return &DLQProducer{
		writer: writer,
		logger: logg,
	}
}

// SendToDLQ отправляет сообщение в DLQ
func (p *DLQProducer) SendToDLQ(ctx context.Context, originalValue []byte, kafkaMsg kafka.Message, errors []string) error {
	if len(errors) == 0 {
		errors = []string{"unknown error"}
	}

	dlqMsg := messages.DeadLetterMessage[json.RawMessage]{
		OriginalMessage: originalValue,
		Topic:           kafkaMsg.Topic,
		Partition:       kafkaMsg.Partition,
		Offset:          kafkaMsg.Offset,
		Key:             string(kafkaMsg.Key),
		Errors:          errors,
		FailedAt:        time.Now(),
		LastError:       errors[len(errors)-1],
	}

	data, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   kafkaMsg.Key,
		Value: data,
		Headers: []kafka.Header{
			{Key: "original-topic", Value: []byte(kafkaMsg.Topic)},
			{Key: "original-partition", Value: []byte(fmt.Sprintf("%d", kafkaMsg.Partition))},
			{Key: "original-offset", Value: []byte(fmt.Sprintf("%d", kafkaMsg.Offset))},
			{Key: "error-count", Value: []byte(fmt.Sprintf("%d", len(errors)))},
			{Key: "failed-at", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Critical: failed to send to DLQ: %v", err))
		return err
	}

	p.logger.Info(fmt.Sprintf("Message sent to DLQ: topic=%s, partition=%d, offset=%d, errors=%d",
		kafkaMsg.Topic, kafkaMsg.Partition, kafkaMsg.Offset, len(errors)))
	return nil
}

func (p *DLQProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
