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

type NotificationMessage struct {
	EventID   string    `json:"eventId"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type NotificationProducer struct {
	kafkaProducer *KafkaProducer
}

func NewNotificationProducer(config config.KafkaConf, topicName string) *NotificationProducer {
	return &NotificationProducer{
		kafkaProducer: NewKafkaProducer(config, topicName),
	}
}

func (p *NotificationProducer) SendMessage(ctx context.Context, message *NotificationMessage) error {
	msg := messages.Message[NotificationMessage]{
		Payload: *message,
	}

	headers := []kafka.Header{
		{Key: "message-id", Value: []byte(uuid.New().String())},
		{Key: "command", Value: []byte("notification")},
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

func (p *NotificationProducer) Close() error {
	if p.kafkaProducer != nil {
		return p.kafkaProducer.Close()
	}
	return nil
}
