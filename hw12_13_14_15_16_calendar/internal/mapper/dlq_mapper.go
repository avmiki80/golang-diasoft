package mapper

import (
	"encoding/json"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/messages"
	"github.com/segmentio/kafka-go"
)

// DLQMessageToDomain конвертирует DeadLetterMessage и kafka.Message в domain.DLQAlert
func DLQMessageToDomain(dlqMsg messages.DeadLetterMessage[json.RawMessage], kafkaMsg kafka.Message) domain.DLQAlert {
	return domain.DLQAlert{
		Topic:           kafkaMsg.Topic,
		OriginalTopic:   dlqMsg.Topic,
		Partition:       dlqMsg.Partition,
		MessageOffset:   dlqMsg.Offset,
		MessageKey:      dlqMsg.Key,
		ErrorMessage:    dlqMsg.LastError,
		ErrorCount:      len(dlqMsg.Errors),
		OriginalMessage: string(dlqMsg.OriginalMessage),
		FailedAt:        dlqMsg.FailedAt,
		Status:          "new",
	}
}
