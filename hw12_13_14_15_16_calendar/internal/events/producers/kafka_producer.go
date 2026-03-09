package producers

import (
	"context"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer создает producer один раз при старте приложения
func NewKafkaProducer(config config.KafkaConf, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.BootstrapServers...),
		Topic:    topic,
		Balancer: getBalancer(config.Producer.Balancer), // или &kafka.Hash{} для партиционирования по ключу

		// Настройки производительности
		BatchSize:    config.Producer.BatchSize,                                      // Размер батча
		BatchTimeout: time.Duration(config.Producer.BatchTimeout) * time.Millisecond, // Таймаут батча

		// Настройки надежности
		RequiredAcks: getRequiredAcks(config.Producer.RequiredAcks), // или RequireAll для максимальной надежности
		MaxAttempts:  config.Producer.MaxAttempts,                   // Количество повторных попыток

		// Асинхронная отправка для лучшей производительности
		Async: config.Producer.Async, // true для fire-and-forget, false для гарантии доставки

		// Compression
		Compression: getCompression(config.Producer.Compression), // или kafka.Gzip, kafka.Lz4
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// SendMessage отправляет сообщение (переиспользует существующее соединение)
func (p *KafkaProducer) SendMessage(ctx context.Context, key, value []byte, headers []kafka.Header) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(),
	})
	return err
}

// Close закрывает writer при завершении работы приложения
func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

func getBalancer(balancerType string) kafka.Balancer {
	switch balancerType {
	case "hash":
		return &kafka.Hash{}
	case "least_bytes":
		return &kafka.LeastBytes{}
	case "random":
		return &kafka.RoundRobin{}
	default:
		return &kafka.Hash{} // по умолчанию используем Hash
	}
}

func getCompression(compressionType string) kafka.Compression {
	switch compressionType {
	case "gzip":
		return kafka.Gzip
	case "snappy":
		return kafka.Snappy
	case "lz4":
		return kafka.Lz4
	case "none":
		return kafka.Compression(0)
	default:
		return kafka.Snappy // по умолчанию используем Snappy
	}
}

func getRequiredAcks(acks int) kafka.RequiredAcks {
	switch acks {
	case -1:
		return kafka.RequireAll
	case 0:
		return kafka.RequireNone
	case 1:
		return kafka.RequireOne
	default:
		return kafka.RequireOne // по умолчанию используем RequireOne
	}
}
