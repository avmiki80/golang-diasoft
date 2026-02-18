package internalhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func NewServerWithGeneratedHandlers(log logger.Logger, eventHandler *handlers.EventHandler, addr string) *ServerNew {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(handlers.LoggingMiddleware(log))

	handlers.RegisterHandlers(e, eventHandler, "")

	e.Server.ReadTimeout = defaultReadTimeout
	e.Server.WriteTimeout = defaultWriteTimeout

	return &ServerNew{
		echo:   e,
		logger: log,
		url:    addr,
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
