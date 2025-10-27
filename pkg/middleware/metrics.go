package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/reqctx"
)

// RequestIDMiddleware adds request ID to context for tracing
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to context
		ctx := reqctx.SetRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		// Add to response headers
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
