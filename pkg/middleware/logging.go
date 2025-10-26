package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/logger"
	"github.com/fanzru/social-media-service-go/pkg/reqctx"
)

// LoggingMiddleware logs all incoming requests with detailed information
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()
			
			// Get request ID from context
			requestID := reqctx.GetRequestID(ctx)
			
			// Read request body
			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
			
			// Extract headers (excluding sensitive ones)
			headers := make(map[string]string)
			for name, values := range r.Header {
				// Skip sensitive headers
				if isSensitiveHeader(name) {
					headers[name] = "[REDACTED]"
				} else {
					headers[name] = strings.Join(values, ", ")
				}
			}
			
			// Parse request body if it's JSON
			var parsedBody interface{}
			if len(requestBody) > 0 && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				json.Unmarshal(requestBody, &parsedBody)
			}
			
			// Log incoming request
			logger.GetGlobal().Info("API Request",
				"requestId", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"headers", headers,
				"body", parsedBody,
				"userAgent", r.UserAgent(),
				"remoteAddr", r.RemoteAddr,
			)
			
			// Create response writer wrapper to capture response
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			// Process request
			next.ServeHTTP(wrapper, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Log response
			logger.GetGlobal().Info("API Response",
				"requestId", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"statusCode", wrapper.statusCode,
				"duration_ms", duration.Milliseconds(),
				"duration_ns", duration.Nanoseconds(),
			)
		})
	}
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

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(name string) bool {
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}
	return sensitiveHeaders[strings.ToLower(name)]
}
