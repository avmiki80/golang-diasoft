package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// RegistryConfig содержит конфигурацию для создания Prometheus registry
type RegistryConfig struct {
	// DB connection для сбора статистики БД (опционально)
	DB *sql.DB
}

// NewPrometheusRegistry создает и настраивает Prometheus registry
// с стандартными коллекторами (Go runtime, process, DB stats)
func NewPrometheusRegistry(cfg *RegistryConfig) (*prometheus.Registry, *Metric) {
	reg := prometheus.NewRegistry()

	// Регистрируем стандартные Go метрики (память, GC, горутины)
	reg.MustRegister(collectors.NewGoCollector())

	// Регистрируем метрики процесса (CPU, память процесса, file descriptors)
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Регистрируем метрики БД, если передано соединение
	if cfg != nil && cfg.DB != nil {
		reg.MustRegister(collectors.NewDBStatsCollector(cfg.DB, "calendar_db"))
	}

	m := NewMetrics(reg)

	return reg, m
}
