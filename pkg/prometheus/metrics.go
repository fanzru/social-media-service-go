package prometheus

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	DBQueriesTotal      *prometheus.CounterVec
	DBQueryDuration     *prometheus.HistogramVec
	ActiveConnections   prometheus.Gauge
	ResponseSize        *prometheus.HistogramVec
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Number of active connections",
			},
		),
		ResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),
	}
}

// HTTPMiddleware creates a middleware for HTTP metrics
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Extract method and endpoint
		method := r.Method
		endpoint := normalizeEndpoint(r.URL.Path)
		status := wrapped.statusCode
		statusStr := getStatusString(status)

		// Record metrics
		m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusStr).Inc()
		m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

		// Record response size if available
		if contentLength := wrapped.Header().Get("Content-Length"); contentLength != "" {
			// Parse content length and record
			if size, err := parseContentLength(contentLength); err == nil {
				m.ResponseSize.WithLabelValues(method, endpoint).Observe(float64(size))
			}
		}
	})
}

// RecordDBQuery records database query metrics
func (m *Metrics) RecordDBQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// Handler returns the Prometheus metrics handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizeEndpoint normalizes the endpoint for metrics
func normalizeEndpoint(path string) string {
	// Remove leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Replace numeric IDs with placeholder
	// This is a simple implementation - you might want to use regex for more complex cases
	if path == "" {
		return "root"
	}

	return path
}

// getStatusString converts status code to string category
func getStatusString(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

// parseContentLength parses content length string to int64
func parseContentLength(contentLength string) (int64, error) {
	// Simple implementation - you might want to use strconv.ParseInt for more robust parsing
	return 0, nil // Placeholder implementation
}
