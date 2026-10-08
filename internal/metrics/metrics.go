package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/chilljzz/gohub/internal/ws"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry *prometheus.Registry

	httpRequestsTotal *prometheus.CounterVec

	httpRequestDurationSeconds *prometheus.HistogramVec
}

func New(manager *ws.Manager) *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry: registry,

		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gohub",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{
				"method",
				"route",
				"status",
			},
		),

		httpRequestDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gohub",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",

				Buckets: []float64{
					0.005,
					0.01,
					0.025,
					0.05,
					0.1,
					0.25,
					0.5,
					1,
					2.5,
					5,
				},
			},
			[]string{
				"method",
				"route",
			},
		),
	}

	websocketConnections := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "gohub",
			Subsystem: "websocket",
			Name:      "connections",
			Help:      "Current WebSocket connections",
		},
		func() float64 {
			return float64(
				manager.ConnectionCount(),
			)
		},
	)

	websocketOnlineUsers := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "gohub",
			Subsystem: "websocket",
			Name:      "online_users",
			Help:      "Current online WebSocket users",
		},

		func() float64 {
			return float64(
				manager.OnlineUserCount(),
			)
		},
	)

	registry.MustRegister(
		collectors.NewGoCollector(),

		collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{},
		),

		m.httpRequestsTotal,

		m.httpRequestDurationSeconds,

		websocketConnections,

		websocketOnlineUsers,
	)

	return m

}

func (
	m *Metrics,
) ObserveHTTPRequest(
	method string,
	route string,
	status int,
	duration time.Duration,
) {
	m.httpRequestsTotal.
		WithLabelValues(
			method,
			route,
			strconv.Itoa(status),
		).
		Inc()

	m.httpRequestDurationSeconds.
		WithLabelValues(
			method,
			route,
		).
		Observe(
			duration.Seconds(),
		)
}

func (
	m *Metrics,
) Handler() http.Handler {

	return promhttp.HandlerFor(
		m.registry,
		promhttp.HandlerOpts{},
	)
}
