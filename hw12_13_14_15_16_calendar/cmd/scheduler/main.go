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

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/app"
	configuration "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/consumers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/events/producers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/memory"
	eventservice "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/robfig/cron/v3"
)

const (
	CalendarNotificationCommandTopic      = "calendar-notification-command"
	CalendarNotificationCommandReplyTopic = "calendar-notification-command-reply"
)

var configFile string

func init() {
	defaultConfig := os.Getenv("CONFIG_FILE")
	if defaultConfig == "" {
		defaultConfig = "./configs/scheduler/config.yaml"
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
	logg.Info("starting scheduler service...")

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

	// Инициализация репозитория
	eventRepo, err := initEventRepository(config.DB, txManager)
	if err != nil {
		return fmt.Errorf("failed to setup event repository: %w", err)
	}

	eventService := eventservice.NewEventService(eventRepo, txManager)

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

	notificationProducer := producers.NewNotificationProducer(*config.KAFKA, CalendarNotificationCommandTopic)

	// Инициализация consumer
	consumer := initConsumer(config, eventService, logg)
	defer consumer.Close()

	notificationApp := app.NewNotificationApp(eventService, notificationProducer, logg)
	defer func() {
		if err := notificationApp.Close(); err != nil {
			logg.Error("failed to close notification app: " + err.Error())
		}
	}()

	// Запуск HTTP сервера для health checks
	httpServer := startHTTPServer(config.HTTP, logg)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logg.Error("failed to shutdown HTTP server: " + err.Error())
		}
	}()

	// Создание cron планировщика
	c := cron.New(cron.WithSeconds())
	defer c.Stop()

	// Добавление задачи отправки уведомлений
	notificationSpec := fmt.Sprintf("@every %ds", config.Scheduler.NotificationInterval)
	_, err = c.AddFunc(notificationSpec, func() {
		logg.Debug("notification cron job triggered")
		if err := notificationApp.SendNotification(ctx); err != nil {
			logg.Error("failed to send notifications: " + err.Error())
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add notification cron job: %w", err)
	}

	// Добавление задачи очистки старых событий
	cleanupSpec := fmt.Sprintf("@every %ds", config.Scheduler.CleanupInterval)
	_, err = c.AddFunc(cleanupSpec, func() {
		logg.Debug("cleanup cron job triggered")
		if err := notificationApp.DeleteOldEvents(ctx); err != nil {
			logg.Error("failed to delete old events: " + err.Error())
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cleanup cron job: %w", err)
	}

	// Запускаем consumer в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- consumer.Run(ctx)
	}()

	// Запуск планировщика
	c.Start()

	logg.Info("scheduler service started successfully")
	logg.Info(fmt.Sprintf("notification interval: %d seconds", config.Scheduler.NotificationInterval))
	logg.Info(fmt.Sprintf("cleanup interval: %d seconds", config.Scheduler.CleanupInterval))

	// Выполнить задачу отправки уведомлений сразу при старте
	go func() {
		if err := notificationApp.SendNotification(ctx); err != nil {
			logg.Error("initial notification send failed: " + err.Error())
		}
	}()

	// Канал для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения
	select {
	case sig := <-sigChan:
		logg.Info(fmt.Sprintf("received signal: %v, shutting down gracefully...", sig))
		cancel()
		return nil
	case err := <-errChan:
		if err != nil {
			logg.Error("consumer error: " + err.Error())
			return err
		}
	case <-ctx.Done():
		logg.Info("context cancelled, shutting down...")
		return nil
	}
	logg.Info("scheduler service stopped")
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
	case "memory":
		logg.Info("using in-memory storage")
		return nil, func() {}, nil

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

func initEventRepository(dbConf configuration.DBConf, txManager database.TxManager) (repositories.CompositeEventRepository, error) {
	switch dbConf.Type {
	case "memory":
		return initMemoryEventRepository()
	case "db":
		return initDBEventRepository(txManager)
	default:
		return nil, fmt.Errorf("unknown database type: %s", dbConf.Type)
	}
}

func initMemoryEventRepository() (repositories.CompositeEventRepository, error) {
	crudRepo := memory.NewEventCrudRepository()
	repo, err := memory.NewEventRepository(crudRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory event repository: %w", err)
	}
	return repo, nil
}

func initDBEventRepository(txManager database.TxManager) (repositories.CompositeEventRepository, error) {
	sqlxDB := txManager.GetDB()
	crudRepo := db.NewEventCrudRepository(sqlxDB)
	repo, err := db.NewEventRepository(crudRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create db event repository: %w", err)
	}
	return repo, nil
}

func initConsumer(config *configuration.Config, eventService eventservice.EventService, logg logger.Logger) consumers.Consumer {
	// Создаем обработчик для подтверждений отправленных уведомлений
	sentNotificationHandler := consumers.NewSentNotificationHandler(eventService, logg)

	// Создаем простой consumer без retry (для обработки подтверждений)
	return consumers.NewKafkaConsumer(
		*config.KAFKA,
		CalendarNotificationCommandReplyTopic,
		sentNotificationHandler,
		logg,
	)
}
