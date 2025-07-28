package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	requestsTotal     *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	responseSize      *prometheus.HistogramVec
	activeConnections prometheus.Gauge
	dbConnections     *prometheus.GaugeVec
}

// New creates a new metrics instance
func New() *Metrics {
	return &Metrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),
		responseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "Size of HTTP responses in bytes",
				Buckets: []float64{200, 500, 900, 1500, 3000, 6000, 12000, 24000, 48000, 96000},
			},
			[]string{"method", "endpoint"},
		),
		activeConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
		dbConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections by state",
			},
			[]string{"state"},
		),
	}
}

// Middleware returns a Gin middleware for collecting HTTP metrics
func (m *Metrics) Middleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Increment active connections
		m.activeConnections.Inc()
		defer m.activeConnections.Dec()

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()

		// If no route matched, use the raw path
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Record metrics
		m.requestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		m.requestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
		m.responseSize.WithLabelValues(method, endpoint).Observe(float64(c.Writer.Size()))
	})
}

// UpdateDBConnections updates database connection metrics
func (m *Metrics) UpdateDBConnections(open, idle, inUse int) {
	m.dbConnections.WithLabelValues("open").Set(float64(open))
	m.dbConnections.WithLabelValues("idle").Set(float64(idle))
	m.dbConnections.WithLabelValues("in_use").Set(float64(inUse))
}

// Custom metrics for business logic
var (
	ConfigTemplatesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "config_templates_total",
			Help: "Total number of configuration templates",
		},
		[]string{"environment", "format", "active"},
	)

	ConfigTemplateOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "config_template_operations_total",
			Help: "Total number of configuration template operations",
		},
		[]string{"operation", "environment", "status"},
	)

	ConfigTemplateSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "config_template_size_bytes",
			Help:    "Size of configuration templates in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"environment", "format"},
	)
)

// RecordTemplateOperation records a template operation metric
func RecordTemplateOperation(operation, environment, status string) {
	ConfigTemplateOperations.WithLabelValues(operation, environment, status).Inc()
}

// RecordTemplateSize records a template size metric
func RecordTemplateSize(environment, format string, size int) {
	ConfigTemplateSize.WithLabelValues(environment, format).Observe(float64(size))
}

// UpdateTemplateCount updates template count metrics
func UpdateTemplateCount(environment, format string, active bool, count int) {
	activeStr := "false"
	if active {
		activeStr = "true"
	}
	ConfigTemplatesTotal.WithLabelValues(environment, format, activeStr).Set(float64(count))
}
