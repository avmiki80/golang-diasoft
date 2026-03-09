package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger    LoggerConf    `toml:"logger" yaml:"logger"`
	HTTP      HTTPConf      `toml:"http" yaml:"http"`
	DB        DBConf        `toml:"database" yaml:"database"`
	KAFKA     *KafkaConf    `toml:"kafka" yaml:"kafka"`
	Scheduler SchedulerConf `toml:"scheduler" yaml:"scheduler"`
}

type TopicConf struct {
	Name              string `toml:"name" yaml:"name"`
	Partitions        int    `toml:"partitions" yaml:"partitions"`
	ReplicationFactor int    `toml:"replication_factor" yaml:"replication_factor"`
}

type KafkaConf struct {
	BootstrapServers []string     `toml:"bootstrap_servers" yaml:"bootstrap_servers"`
	TopicConfigs     []TopicConf  `toml:"topic_configs" yaml:"topic_configs"`
	Producer         ProducerConf `toml:"producer" yaml:"producer"`
	Consumer         ConsumerConf `toml:"consumer" yaml:"consumer"`
}

type ProducerConf struct {
	Balancer     string `toml:"balancer" yaml:"balancer"`
	RequiredAcks int    `toml:"required_acks" yaml:"required_acks"`
	MaxAttempts  int    `toml:"max_attempts" yaml:"max_attempts"`
	BatchSize    int    `toml:"batch_size" yaml:"batch_size"`
	BatchTimeout int    `toml:"batch_timeout" yaml:"batch_timeout"`
	Compression  string `toml:"compression" yaml:"compression"`
	Async        bool   `toml:"async" yaml:"async"`
}

type ConsumerConf struct {
	GroupID         string  `toml:"group_id" yaml:"group_id"`
	MaxAttempts     int     `toml:"max_attempts" yaml:"max_attempts"`
	InitialInterval int     `toml:"initial_interval" yaml:"initial_interval"`
	MaxInterval     int     `toml:"max_interval" yaml:"max_interval"`
	Multiplier      float64 `toml:"multiplier" yaml:"multiplier"`
	MaxElapsedTime  int     `toml:"max_elapsed_time" yaml:"max_elapsed_time"`
	RetryEnable     bool    `toml:"retry_enable" yaml:"retry_enable"`
	AutoCommit      *bool   `toml:"auto_commit" yaml:"auto_commit"`
}

type SchedulerConf struct {
	NotificationInterval int `toml:"notification_interval" yaml:"notification_interval"`
	CleanupInterval      int `toml:"cleanup_interval" yaml:"cleanup_interval"`
}

type LoggerConf struct {
	Level string `toml:"level" yaml:"level"`
}

type HTTPConf struct {
	Host string `toml:"host" yaml:"host"`
	Port string `toml:"port" yaml:"port"`
}

type DBConf struct {
	Type string `toml:"type" yaml:"type"`
	DSN  string `toml:"dsn" yaml:"dsn"`
}

func NewConfig(path string) (*Config, error) {
	confData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		//nolint
		if err = yaml.Unmarshal(confData, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml config: %w", err)
		}
	case ".toml":
		if err = toml.Unmarshal(confData, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal toml config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s (supported: .yaml, .yml, .toml)", ext)
	}

	if config.Logger.Level == "" {
		config.Logger.Level = "INFO"
	}
	if config.HTTP.Host == "" {
		config.HTTP.Host = "localhost"
	}
	if config.HTTP.Port == "" {
		config.HTTP.Port = "8080"
	}
	if config.DB.Type == "" {
		config.DB.Type = "memory"
	}
	if config.KAFKA != nil {
		setKafkaDefaults(config.KAFKA)
	}

	if config.Scheduler.NotificationInterval == 0 {
		config.Scheduler.NotificationInterval = 60
	}
	if config.Scheduler.CleanupInterval == 0 {
		config.Scheduler.CleanupInterval = 86400
	}

	return &config, nil
}

func setBalancer(balancer string) string {
	switch balancer {
	case "hash", "least_bytes", "random":
		return balancer
	default:
		return "hash"
	}
}

func setCompression(compression string) string {
	switch compression {
	case "none", "gzip", "snappy", "lz4":
		return compression
	default:
		return "snappy"
	}
}

func setKafkaDefaults(kafka *KafkaConf) {
	if len(kafka.BootstrapServers) == 0 {
		kafka.BootstrapServers = []string{"localhost:9092"}
	}

	// Установка дефолтных значений для топиков
	for i := range kafka.TopicConfigs {
		if kafka.TopicConfigs[i].Partitions == 0 {
			kafka.TopicConfigs[i].Partitions = 3
		}
		if kafka.TopicConfigs[i].ReplicationFactor == 0 {
			kafka.TopicConfigs[i].ReplicationFactor = 1
		}
	}

	if kafka.Consumer.GroupID == "" {
		kafka.Consumer.GroupID = "calendar_notification_group"
	}
	if kafka.Consumer.MaxAttempts == 0 {
		kafka.Consumer.MaxAttempts = 3
	}
	if kafka.Consumer.InitialInterval == 0 {
		kafka.Consumer.InitialInterval = 1
	}
	if kafka.Consumer.MaxInterval == 0 {
		kafka.Consumer.MaxInterval = 10
	}
	if kafka.Consumer.Multiplier == 0 {
		kafka.Consumer.Multiplier = 2.0
	}
	if kafka.Consumer.MaxElapsedTime == 0 {
		kafka.Consumer.MaxElapsedTime = 30
	}
	if kafka.Consumer.AutoCommit == nil {
		defaultAutoCommit := true
		kafka.Consumer.AutoCommit = &defaultAutoCommit
	}

	kafka.Producer.Balancer = setBalancer(kafka.Producer.Balancer)
	if kafka.Producer.RequiredAcks > 1 || kafka.Producer.RequiredAcks < -1 {
		kafka.Producer.RequiredAcks = 0
	}
	if kafka.Producer.MaxAttempts == 0 {
		kafka.Producer.MaxAttempts = 3
	}
	if kafka.Producer.BatchSize == 0 {
		kafka.Producer.BatchSize = 100
	}
	if kafka.Producer.BatchTimeout == 0 {
		kafka.Producer.BatchTimeout = 10
	}
	kafka.Producer.Compression = setCompression(kafka.Producer.Compression)
}

func (c *Config) GetConfig() Config {
	return *c
}

func (c *Config) GetLoggerConfig() LoggerConf {
	return c.Logger
}

func (c *Config) GetHTTPConfig() HTTPConf {
	return c.HTTP
}

func (c *Config) GetDBConfig() DBConf {
	return c.DB
}

func (c *Config) GetKafkaConfig() KafkaConf {
	return *c.KAFKA
}
