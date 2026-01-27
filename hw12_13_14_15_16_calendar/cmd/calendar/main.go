package main

import (
	"context"
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
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/db"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/repositories/memory"
	internalhttp "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers"
	eventservice "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/services"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
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

	logg := logger.New(config.Logger.Level)

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
	calendar := app.New(eventService, notifyService, logg)

	server := initHTTPServer(config.HTTP, calendar, logg)

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

// TODO: Примеры создания других репозиториев:
//
// func setupNotificationRepository(dbConf configuration.DBConf, txManager database.TxManager, logg logger.Logger) (repositories.NotificationRepository, error) {
//     switch dbConf.Type {
//     case "memory":
//         return setupMemoryNotificationRepository(logg)
//     case "db":
//         return setupDBNotificationRepository(txManager, logg)
//     default:
//         return nil, fmt.Errorf("unknown database type: %s", dbConf.Type)
//     }
// }
//
// func setupUserRepository(dbConf configuration.DBConf, txManager database.TxManager, logg logger.Logger) (repositories.UserRepository, error) {
//     switch dbConf.Type {
//     case "memory":
//         return setupMemoryUserRepository(logg)
//     case "db":
//         return setupDBUserRepository(txManager, logg)
//     default:
//         return nil, fmt.Errorf("unknown database type: %s", dbConf.Type)
//     }
// }

func initHTTPServer(httpConf configuration.HTTPConf, calendar *app.App, logg logger.Logger) *internalhttp.Server {
	eventHandler := handlers.NewEventHandler(calendar, logg)
	serverAddr := httpConf.Host + ":" + httpConf.Port
	return internalhttp.NewServer(logg, eventHandler, serverAddr)
}

func runHTTPServer(server *internalhttp.Server, logg logger.Logger) error {
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
