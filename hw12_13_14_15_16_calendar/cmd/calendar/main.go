package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/app"
	configuration "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/config"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/database"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/metrics"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/memory"
	internalhttp "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers"
	eventservice "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
	"github.com/prometheus/client_golang/prometheus"
)

var configFile string

func init() {
	defaultConfig := os.Getenv("CONFIG_FILE")
	if defaultConfig == "" {
		defaultConfig = "./configs/calendar/config.yaml"
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

	if err := run(config, logg); err != nil {
		logg.Error("application error: " + err.Error())
		os.Exit(1)
	}
}

func run(config *configuration.Config, logg logger.Logger) error {
	txManager, cleanup, err := initDatabase(config.DB, logg)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer cleanup()

	eventRepo, err := initEventRepository(config.DB, txManager)
	if err != nil {
		return fmt.Errorf("failed to setup event repository: %w", err)
	}

	eventService := eventservice.NewEventService(eventRepo, txManager)
	var notifyService eventservice.NotificationService
	calendar := app.NewCalendarApp(eventService, notifyService, logg)

	var sqlDB *sql.DB
	if txManager != nil {
		sqlDB = txManager.GetDB().DB
	}
	reg, m := initMetrics(sqlDB)

	server := initHTTPServer(config.HTTP, calendar, logg, reg, m)

	return runHTTPServer(server, logg)
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

func initMetrics(sqlDB *sql.DB) (*prometheus.Registry, *metrics.Metric) {
	reg, m := metrics.NewPrometheusRegistry(&metrics.RegistryConfig{
		DB: sqlDB,
	})
	return reg, m
}

func initHTTPServer(httpConf configuration.HTTPConf, calendar *app.CalendarApp, logg logger.Logger, reg *prometheus.Registry, m *metrics.Metric) *internalhttp.ServerNew {
	eventHandler := handlers.NewEventHandler(calendar, logg)
	serverAddr := httpConf.Host + ":" + httpConf.Port

	return internalhttp.NewServerWithGeneratedHandlers(&internalhttp.ServerConfig{
		Logger:       logg,
		EventHandler: eventHandler,
		Metric:       m,
		Registry:     reg,
		Addr:         serverAddr,
	})
}

func runHTTPServer(server *internalhttp.ServerNew, logg logger.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		logg.Info("shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
