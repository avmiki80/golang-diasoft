package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	genhandlers "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers/generated"
	"github.com/labstack/echo/v4"
)

func RegisterHandlers(router genhandlers.EchoRouter, handler *EventHandler, url string) {
	if url == "" {
		genhandlers.RegisterHandlers(router, handler)
	} else {
		genhandlers.RegisterHandlersWithBaseURL(router, handler, url)
	}
}

func LoggingMiddleware(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			err := next(c)

			logHTTPRequest(log, req, c.Response().Status, start)

			if err != nil {
				log.Error("Handler error: " + err.Error())
			}

			return err
		}
	}
}

func logHTTPRequest(log logger.Logger, r *http.Request, status int, start time.Time) {
	if status == 0 {
		status = http.StatusOK
	}

	logLine := fmt.Sprintf("%s [%s] %s %s %s %d %v \"%s\"",
		getClientIP(r),
		start.Format("02/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.URL.RequestURI(),
		r.Proto,
		status,
		time.Since(start),
		getUserAgent(r),
	)

	log.Info(logLine)
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}

	return ip
}

func getUserAgent(r *http.Request) string {
	if ua := r.UserAgent(); ua != "" {
		return ua
	}
	return "-"
}
