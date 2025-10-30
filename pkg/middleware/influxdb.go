package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

			// Extract entity and normalize dynamic segments
			rawPath := r.URL.Path
			entity := extractEntity(rawPath)
			normPath := normalizePath(rawPath)

			// Record metrics to InfluxDB
			if influxClient != nil {
				tags := map[string]string{
					"group":       "API_IN",
					"entity":      entity,
					"path":        normPath, // normalized: replace numeric/UUID segments with {id}
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

// normalizePath replaces path segments that look like IDs with {id}
func normalizePath(p string) string {
	if p == "" || p == "/" {
		return p
	}
	// UUID matcher (case-insensitive)
	uuidRe := regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if seg == "" { // leading slash yields empty segment
			continue
		}
		// numeric id
		if _, err := strconv.ParseInt(seg, 10, 64); err == nil {
			parts[i] = "{id}"
			continue
		}
		// uuid id
		if uuidRe.MatchString(seg) {
			parts[i] = "{id}"
			continue
		}
	}
	// Avoid duplicating slashes
	return strings.ReplaceAll(strings.Join(parts, "/"), "//", "/")
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
