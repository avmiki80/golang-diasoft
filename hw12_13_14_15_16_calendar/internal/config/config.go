package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger LoggerConf `toml:"logger" yaml:"logger"`
	HTTP   HTTPConf   `toml:"http" yaml:"http"`
	DB     DBConf     `toml:"database" yaml:"database"`
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

	// Установка значений по умолчанию
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

	return &config, nil
}
