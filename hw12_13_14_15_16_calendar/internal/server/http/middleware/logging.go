package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				status:         0,
			}

			next.ServeHTTP(rw, r)

			logHTTPRequest(log, r, rw.status, start)
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
