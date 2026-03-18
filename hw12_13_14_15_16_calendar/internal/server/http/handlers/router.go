package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/metrics"
	genhandlers "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers/generated"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus"
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

func PrometheusMiddleware(metric metrics.Metric) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}
			method := req.Method

			err := next(c)

			duration := time.Since(start).Seconds()
			status := fmt.Sprintf("%d", c.Response().Status)

			metric.AllRequest.With(prometheus.Labels{
				"method": method,
				"path":   path,
				"status": status,
			}).Inc()

			metric.RequestDuration.With(prometheus.Labels{
				"method": method,
				"path":   path,
			}).Observe(duration)

			if c.Response().Status >= 200 && c.Response().Status < 300 {
				metric.CorrectRequest.With(prometheus.Labels{
					"method": method,
					"path":   path,
				}).Inc()
			} else {
				metric.ErrorRequest.With(prometheus.Labels{
					"method": method,
					"path":   path,
					"status": status,
				}).Inc()
			}

			if err != nil {
				log.Error("Handler error: " + err.Error())
			}

			return err
		}
	}
}

func HealthHandler(metric *metrics.Metric) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := "UP"
		httpStatus := http.StatusOK
		metric.ServiceHealth.Set(1)
		return c.JSON(httpStatus, map[string]any{
			"status":    status,
			"timestamp": time.Now().Format(time.RFC3339),
		})
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
