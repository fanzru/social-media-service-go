package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/influxdb"
)

// InfluxDBMiddleware creates an InfluxDB middleware for HTTP requests
func InfluxDBMiddleware(influxClient *influxdb.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &influxResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process the request
			next.ServeHTTP(wrapper, r)

			// Calculate duration
			duration := time.Since(start)

			// Extract entity from path
			entity := extractEntity(r.URL.Path)

			// Record metrics to InfluxDB
			if influxClient != nil {
				tags := map[string]string{
					"group":       "API_IN",
					"entity":      entity,
					"path":        r.URL.Path, // Full API path untuk tracking detail
					"method":      r.Method,
					"http_status": strconv.Itoa(wrapper.statusCode),
					"code":        getErrorCode(wrapper.statusCode),
				}

				// Record request count
				_ = influxClient.WriteCounter("http_requests_total", tags, 1)

				// Record response time
				_ = influxClient.WriteTiming("http_request_duration_ms", tags, duration)
			}
		})
	}
}

// influxResponseWriter wraps http.ResponseWriter to capture status code
type influxResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *influxResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// extractEntity extracts entity name from URL path
func extractEntity(path string) string {
	// Simple entity extraction logic
	switch {
	case len(path) > 1:
		// Remove leading slash and get first segment
		if path[0] == '/' {
			path = path[1:]
		}
		// Find first slash to get entity
		for i, char := range path {
			if char == '/' {
				return path[:i]
			}
		}
		return path
	default:
		return "unknown"
	}
}

// getErrorCode returns error code based on HTTP status
func getErrorCode(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "SUCCESS"
	case statusCode >= 400 && statusCode < 500:
		return "CLIENT_ERROR"
	case statusCode >= 500:
		return "SERVER_ERROR"
	default:
		return "UNKNOWN"
	}
}
