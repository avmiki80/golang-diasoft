package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metric struct {
	AllRequest     *prometheus.CounterVec
	CorrectRequest *prometheus.CounterVec
	ErrorRequest   *prometheus.CounterVec

	RequestDuration *prometheus.HistogramVec

	ServiceHealth prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metric {
	m := &Metric{
		AllRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total count of HTTP requests to calendar service.",
			},
			[]string{"method", "path", "status"},
		),
		CorrectRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_success_total",
				Help: "Total count of successful HTTP requests (2xx status codes).",
			},
			[]string{"method", "path"},
		),
		ErrorRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_error_total",
				Help: "Total count of failed HTTP requests (non-2xx status codes).",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds.",
				Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
			},
			[]string{"method", "path"},
		),
		ServiceHealth: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "service_health",
				Help: "Health status of calendar service (1 = UP, 0 = DOWN).",
			},
		),
	}

	reg.MustRegister(m.AllRequest)
	reg.MustRegister(m.CorrectRequest)
	reg.MustRegister(m.ErrorRequest)
	reg.MustRegister(m.RequestDuration)
	reg.MustRegister(m.ServiceHealth)

	m.ServiceHealth.Set(1)

	return m
}
