package reqctx

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDKey is the key used to store request ID in context
type RequestIDKey struct{}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}

// SetRequestID sets request ID in context
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey{}, requestID)
}

// ExtractRequestIDFromHeader extracts request ID from HTTP header
func ExtractRequestIDFromHeader(r *http.Request) string {
	// Try X-Request-Id header first
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}

	// Try RequestID header as fallback
	if requestID := r.Header.Get("RequestID"); requestID != "" {
		return requestID
	}

	// Generate new request ID if none provided
	return generateRequestID()
}

// Middleware creates a middleware that extracts request ID and adds it to context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := ExtractRequestIDFromHeader(r)
		ctx := SetRequestID(r.Context(), requestID)

		// Add request ID to response header for tracing
		w.Header().Set("X-Request-Id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return uuid.New().String()
}
