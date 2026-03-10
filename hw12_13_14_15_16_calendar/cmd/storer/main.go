package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	configuration "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/consumers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/producers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
)

const (
	CalendarNotificationCommandTopic      = "calendar-notification-command"
	CalendarNotificationCommandReplyTopic = "calendar-notification-command-reply"
)

var configFile string

func init() {
	defaultConfig := os.Getenv("CONFIG_FILE")
	if defaultConfig == "" {
		defaultConfig = "./configs/storer/config.yaml"
	}

	flag.StringVar(&configFile, "config", defaultConfig, "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := configuration.NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(config.Logger.Level, os.Stdout)
	logg.Info("starting storer service...")

	if err := run(config, logg); err != nil {
		logg.Error("application error: " + err.Error())
		os.Exit(1)
	}
}

func run(config *configuration.Config, logg logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация БД
	txManager, dbCleanup, err := initDatabase(config.DB, logg)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer dbCleanup()

	// Проверка наличия топиков в конфигурации
	if err := events.ValidateTopicExists(config.KAFKA, CalendarNotificationCommandTopic); err != nil {
		logg.Error(err.Error())
		return err
	}
	if err := events.ValidateTopicExists(config.KAFKA, CalendarNotificationCommandReplyTopic); err != nil {
		logg.Error(err.Error())
		return err
	}

	// Создать топики если их нет
	if err := events.EnsureTopicsExist(config.KAFKA, logg); err != nil {
		return fmt.Errorf("failed to ensure topics exist: %w", err)
	}

	// Инициализация репозитория уведомлений
	notificationRepo := initDBNotificationRepository(txManager)

	// Инициализация сервиса уведомлений
	notificationService := initNotificationService(notificationRepo, txManager)

	// Инициализация consumer
	consumer := initConsumer(config, notificationService, logg)
	defer consumer.Close()

	// Запуск HTTP сервера для health checks
	httpServer := startHTTPServer(config.HTTP, logg)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logg.Error("failed to shutdown HTTP server: " + err.Error())
		}
	}()

	// Канал для сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем consumer в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- consumer.Run(ctx)
	}()

	logg.Info("storer service started successfully")

	// Ждем сигнал завершения или ошибку
	select {
	case sig := <-sigChan:
		logg.Info(fmt.Sprintf("received signal: %v, shutting down gracefully...", sig))
		cancel()
		time.Sleep(2 * time.Second) // Даем время на graceful shutdown
	case err := <-errChan:
		if err != nil {
			logg.Error("consumer error: " + err.Error())
			return err
		}
	case <-ctx.Done():
		logg.Info("context cancelled, shutting down...")
	}

	logg.Info("storer service stopped")
	return nil
}

func startHTTPServer(httpConf configuration.HTTPConf, logg logger.Logger) *http.Server {
	mux := http.NewServeMux()

	//// Health check endpoint
	//mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("OK"))
	//})
	//
	//// Readiness check endpoint
	//mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("READY"))
	//})

	addr := httpConf.Host + ":" + httpConf.Port
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		logg.Info("HTTP server listening on " + addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("HTTP server error: " + err.Error())
		}
	}()

	return server
}

type cleanupFunc func()

func initDatabase(dbConf configuration.DBConf, logg logger.Logger) (database.TxManager, cleanupFunc, error) {
	switch dbConf.Type {
	case "db":
		logg.Info("connecting to PostgreSQL database...")
		poolConfig := database.DefaultConnectionConfig(dbConf.DSN)
		sqlxDB, err := database.NewConnection(poolConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		logg.Info("database connection established")

		txManager := database.NewTxManager(sqlxDB)

		cleanup := func() {
			if err := sqlxDB.Close(); err != nil {
				logg.Error("failed to close database connection: " + err.Error())
			}
		}

		return txManager, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("unknown database type: %s", dbConf.Type)
	}
}

func initDBNotificationRepository(txManager database.TxManager) repositories.CompositeNotificationRepository {
	sqlxDB := txManager.GetDB()
	crudRepo := db.NewNotificationCrudRepository(sqlxDB)
	repo, err := db.NewNotificationRepository(crudRepo)
	if err != nil {
		panic(fmt.Sprintf("failed to create notification repository: %v", err))
	}
	return repo
}

func initNotificationService(repo repositories.CompositeNotificationRepository, txManager database.TxManager) services.NotificationService {
	return services.NewNotificationService(repo, txManager)
}

func initConsumer(config *configuration.Config, notificationService services.NotificationService, logg logger.Logger) consumers.Consumer {
	// Создаем producer для отправки подтверждений
	sentNotificationProducer := producers.NewSentNotificationProducer(*config.KAFKA, CalendarNotificationCommandReplyTopic)

	// Создаем обработчик уведомлений
	notificationHandler := consumers.NewNotificationHandler(notificationService, sentNotificationProducer, logg)

	// Создаем DLQ producer
	dlqProducer := producers.NewDLQProducer(config.KAFKA.BootstrapServers, CalendarNotificationCommandTopic, logg)

	// Создаем retryable consumer с DLQ поддержкой
	return consumers.NewRetryableKafkaConsumer(
		*config.KAFKA,
		CalendarNotificationCommandTopic,
		notificationHandler,
		dlqProducer,
		logg,
	)
}
