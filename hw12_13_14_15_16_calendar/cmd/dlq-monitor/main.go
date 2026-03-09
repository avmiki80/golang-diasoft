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
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const CalendarNotificationCommandDLQTopic = "calendar-notification-command-dlq"

var configFile string

func init() {
	defaultConfig := os.Getenv("CONFIG_FILE")
	if defaultConfig == "" {
		defaultConfig = "./configs/dlq-monitor/config.prod.yaml"
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
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(config.Logger.Level, os.Stdout)
	logg.Info("starting DLQ Monitor service...")

	if err := run(config, logg); err != nil {
		logg.Error("Application error: " + err.Error())
		os.Exit(1)
	}
}

func run(config *configuration.Config, logg logger.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Инициализация базы данных
	txManager, cleanup, err := initDatabase(config.DB, logg)
	if err != nil {
		return err
	}
	defer cleanup()

	// Проверка наличия DLQ топика в конфигурации
	if err := events.ValidateTopicExists(config.KAFKA, CalendarNotificationCommandDLQTopic); err != nil {
		logg.Error(err.Error())
		return err
	}

	// Создать топики если их нет
	if err := events.EnsureTopicsExist(config.KAFKA, logg); err != nil {
		return fmt.Errorf("failed to ensure topics exist: %w", err)
	}

	// Инициализация репозитория и сервиса для DLQ алертов
	dlqAlertRepo := initDLQAlertRepository(txManager)
	dlqAlertService := initDLQAlertService(dlqAlertRepo, txManager)

	// Инициализация consumer
	consumer := initConsumer(config, dlqAlertService, logg)
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

	// Запускаем consumer в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- consumer.Run(ctx)
	}()

	logg.Info("DLQ Monitor service started")

	// Ожидание сигнала завершения или ошибки
	select {
	case err := <-errChan:
		if err != nil {
			logg.Error("consumer error: " + err.Error())
			return err
		}
	case <-ctx.Done():
		logg.Info("context cancelled, shutting down...")
		return nil
	}

	logg.Info("DLQ Monitor service stopped")
	return nil
}

func startHTTPServer(httpConf configuration.HTTPConf, logg logger.Logger) *http.Server {
	mux := http.NewServeMux()

	// Health check endpoint
	//mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("OK"))
	//})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", httpConf.Host, httpConf.Port),
		Handler: mux,
	}

	go func() {
		logg.Info(fmt.Sprintf("HTTP server listening on %s:%s", httpConf.Host, httpConf.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("HTTP server error: " + err.Error())
		}
	}()

	return server
}

func initDatabase(dbConf configuration.DBConf, logg logger.Logger) (database.TxManager, func(), error) {
	switch dbConf.Type {
	case "db":
		sqlxDB, err := sqlx.Connect("postgres", dbConf.DSN)
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

func initDLQAlertRepository(txManager database.TxManager) repositories.CompositeDLQAlertRepository {
	sqlxDB := txManager.GetDB()
	return db.NewDLQAlertCrudRepository(sqlxDB)
}

func initDLQAlertService(repo repositories.CompositeDLQAlertRepository, txManager database.TxManager) services.DLQAlertService {
	return services.NewDLQAlertService(repo, txManager)
}

func initConsumer(config *configuration.Config, dlqAlertService services.DLQAlertService, logg logger.Logger) consumers.Consumer {
	// Создаем обработчик для DLQ сообщений
	dlqMonitorHandler := consumers.NewDLQMonitorHandler(dlqAlertService, logg)

	// Создаем простой consumer без retry (для мониторинга DLQ)
	return consumers.NewKafkaConsumer(
		*config.KAFKA,
		CalendarNotificationCommandDLQTopic,
		dlqMonitorHandler,
		logg,
	)
}
