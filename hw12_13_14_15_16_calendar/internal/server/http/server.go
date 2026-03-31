package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/metrics"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

type ServerNew struct {
	echo   *echo.Echo
	logger logger.Logger
	url    string
}

// ServerConfig содержит конфигурацию для создания HTTP сервера
type ServerConfig struct {
	Logger       logger.Logger
	EventHandler *handlers.EventHandler
	Metric       *metrics.Metric
	Registry     *prometheus.Registry
	Addr         string
}

func NewServerWithGeneratedHandlers(config *ServerConfig) *ServerNew {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(handlers.LoggingMiddleware(config.Logger))
	e.Use(handlers.PrometheusMiddleware(*config.Metric))

	if config.EventHandler != nil {
		handlers.RegisterHandlers(e, config.EventHandler, "")
	}

	// Регистрируем эндпоинт для метрик Prometheus
	e.GET("/actuator/prometheus", echo.WrapHandler(promhttp.HandlerFor(config.Registry, promhttp.HandlerOpts{Registry: config.Registry})))

	// Регистрируем health check эндпоинт
	e.GET("/health", handlers.HealthHandler(config.Metric))

	e.Server.ReadTimeout = defaultReadTimeout
	e.Server.WriteTimeout = defaultWriteTimeout

	return &ServerNew{
		echo:   e,
		logger: config.Logger,
		url:    config.Addr,
	}
}

func (s *ServerNew) Start(ctx context.Context) error {
	s.logger.Info("starting HTTP server")

	errChan := make(chan error, 1)
	go func() {
		if err := s.echo.Start(s.url); err != nil && errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		s.logger.Error("server error: " + err.Error())
		return err
	case <-ctx.Done():
		s.logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		return s.echo.Shutdown(shutdownCtx)
	}
}

func (s *ServerNew) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	return s.echo.Shutdown(ctx)
}
