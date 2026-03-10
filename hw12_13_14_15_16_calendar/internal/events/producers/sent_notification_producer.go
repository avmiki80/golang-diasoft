package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/messages"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type SentNotificationMessage struct {
	EventID string `json:"eventId"`
	UserID  string `json:"userId"`
}

type SentNotificationProducer struct {
	kafkaProducer *KafkaProducer
}

func NewSentNotificationProducer(config config.KafkaConf, topicName string) *SentNotificationProducer {
	return &SentNotificationProducer{
		kafkaProducer: NewKafkaProducer(config, topicName),
	}
}

func (p *SentNotificationProducer) SendMessage(ctx context.Context, message *SentNotificationMessage) error {
	msg := messages.Message[SentNotificationMessage]{
		Payload: *message,
	}

	headers := []kafka.Header{
		{Key: "message-id", Value: []byte(uuid.New().String())},
		{Key: "command", Value: []byte("sent-notification")},
		{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339Nano))},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.kafkaProducer.SendMessage(ctx, []byte(msg.Payload.UserID), data, headers)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (p *SentNotificationProducer) Close() error {
	if p.kafkaProducer != nil {
		return p.kafkaProducer.Close()
	}
	return nil
}
