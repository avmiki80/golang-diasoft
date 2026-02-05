package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/middleware"
	"github.com/gorilla/mux"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
	shutdownTimeout     = 3 * time.Second
)

type Server struct {
	httpServer   *http.Server
	logger       logger.Logger
	eventHandler *handlers.EventHandler
}

func NewServer(log logger.Logger, eventHandler *handlers.EventHandler, addr string) *Server {
	srv := &Server{
		logger:       log,
		eventHandler: eventHandler,
	}

	srv.httpServer = &http.Server{
		Addr:         addr,
		Handler:      srv.setupRouter(),
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	return srv
}

func (s *Server) setupRouter() http.Handler {
	router := mux.NewRouter()

	s.eventHandler.RegisterRoutes(router)

	router.Use(middleware.LoggingMiddleware(s.logger))

	return router
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("starting HTTP server on %s", s.httpServer.Addr))

	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		s.logger.Error(fmt.Sprintf("server error: %v", err))
		return err
	case <-ctx.Done():
		s.logger.Info("shutdown signal received")
		return s.shutdown()
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(fmt.Sprintf("server shutdown error: %v", err))
		return err
	}
	s.logger.Info("HTTP server stopped successfully")
	return nil
}

func (s *Server) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.Stop(ctx)
}
