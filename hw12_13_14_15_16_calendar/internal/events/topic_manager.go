package events

import (
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/segmentio/kafka-go"
)

// ValidateTopicExists проверяет наличие топика в конфигурации Kafka.
// Возвращает nil если топик найден, иначе возвращает ошибку.
func ValidateTopicExists(kafkaConfig *config.KafkaConf, topicName string) error {
	if kafkaConfig == nil {
		return fmt.Errorf("kafka config is nil")
	}
	// Проверяем в TopicConfigs (новый формат)
	for _, topicConf := range kafkaConfig.TopicConfigs {
		if topicConf.Name == topicName {
			return nil
		}
	}
	return fmt.Errorf("topic %s not found in config", topicName)
}

// EnsureTopicsExist создает топики если они не существуют
func EnsureTopicsExist(kafkaConf *config.KafkaConf, logg logger.Logger) error {
	if kafkaConf == nil || len(kafkaConf.TopicConfigs) == 0 {
		logg.Info("No topic configs found, skipping topic creation")
		return nil
	}

	logg.Info(fmt.Sprintf("Connecting to Kafka at %v to ensure topics exist", kafkaConf.BootstrapServers))

	conn, err := kafka.Dial("tcp", kafkaConf.BootstrapServers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	// Получить список существующих топиков
	partitions, err := controllerConn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	existingTopics := make(map[string]bool)
	for _, p := range partitions {
		existingTopics[p.Topic] = true
	}

	// Создать только те топики, которых нет
	var topicsToCreate []kafka.TopicConfig
	for _, topicConf := range kafkaConf.TopicConfigs {
		if !existingTopics[topicConf.Name] {
			topicsToCreate = append(topicsToCreate, kafka.TopicConfig{
				Topic:             topicConf.Name,
				NumPartitions:     topicConf.Partitions,
				ReplicationFactor: topicConf.ReplicationFactor,
			})
			logg.Info(fmt.Sprintf("Will create topic: %s (partitions: %d, replication: %d)",
				topicConf.Name, topicConf.Partitions, topicConf.ReplicationFactor))
		} else {
			logg.Info(fmt.Sprintf("Topic already exists: %s", topicConf.Name))
		}
	}

	if len(topicsToCreate) > 0 {
		err = controllerConn.CreateTopics(topicsToCreate...)
		if err != nil {
			return fmt.Errorf("failed to create topics: %w", err)
		}
		logg.Info(fmt.Sprintf("Successfully created %d topics", len(topicsToCreate)))

		// Подождать немного, чтобы топики были готовы
		time.Sleep(2 * time.Second)
	} else {
		logg.Info("All topics already exist, no need to create")
	}

	return nil
}
