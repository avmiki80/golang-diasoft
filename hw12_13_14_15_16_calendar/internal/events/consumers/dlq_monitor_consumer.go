package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/messages"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/mapper"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/segmentio/kafka-go"
)

// DLQMonitorHandler обрабатывает сообщения из DLQ и сохраняет алерты в БД
type DLQMonitorHandler struct {
	alertService services.DLQAlertService
	logger       logger.Logger
}

// NewDLQMonitorHandler создает новый обработчик DLQ мониторинга
func NewDLQMonitorHandler(
	alertService services.DLQAlertService,
	logg logger.Logger,
) *DLQMonitorHandler {
	return &DLQMonitorHandler{
		alertService: alertService,
		logger:       logg,
	}
}

// Handle обрабатывает сообщение из DLQ
func (h *DLQMonitorHandler) Handle(ctx context.Context, msg kafka.Message) error {
	// Парсим DLQ сообщение
	var dlqMsg messages.DeadLetterMessage[json.RawMessage]
	if err := json.Unmarshal(msg.Value, &dlqMsg); err != nil {
		return fmt.Errorf("failed to unmarshal DLQ message: %w", err)
	}

	h.logger.Info(fmt.Sprintf("DLQ Alert: topic=%s, partition=%d, offset=%d, errors=%d",
		dlqMsg.Topic, dlqMsg.Partition, dlqMsg.Offset, len(dlqMsg.Errors)))

	// Конвертируем в domain.DLQAlert используя маппер
	alert := mapper.DLQMessageToDomain(dlqMsg, msg)

	// Сохраняем в БД
	createdAlert, err := h.alertService.CreateAlert(ctx, alert)
	if err != nil {
		return fmt.Errorf("failed to create DLQ alert: %w", err)
	}

	h.logger.Info(fmt.Sprintf("DLQ Alert saved to database: ID=%s, OriginalTopic=%s, ErrorCount=%d",
		createdAlert.ID, createdAlert.OriginalTopic, createdAlert.ErrorCount))

	return nil
}

// GetName возвращает имя обработчика
func (h *DLQMonitorHandler) GetName() string {
	return "DLQMonitorHandler"
}
