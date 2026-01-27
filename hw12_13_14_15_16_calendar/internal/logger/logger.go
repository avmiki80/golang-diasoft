package logger

import (
	"io"
	"log"
	"strings"
)

// Logger интерфейс для логирования в приложении.
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	Prefix = "calendar: "
)

type logger struct {
	level  LogLevel
	logger *log.Logger
}

func New(level string, writer io.Writer) Logger {
	logLevel := parseLevel(level)
	return &logger{
		level:  logLevel,
		logger: log.New(writer, Prefix, log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func parseLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *logger) Debug(msg string) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] %s", msg)
	}
}

func (l *logger) Info(msg string) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] %s", msg)
	}
}

func (l *logger) Warn(msg string) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] %s", msg)
	}
}

func (l *logger) Error(msg string) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] %s", msg)
	}
}
