package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/social-media-service-go/internal/app/health"
	"github.com/fanzru/social-media-service-go/internal/app/health/port/genhttp"
	"github.com/fanzru/social-media-service-go/pkg/logger"
	"github.com/fanzru/social-media-service-go/pkg/reqctx"
)

// Handler handles HTTP requests for health endpoints
type Handler struct {
	service health.HealthService
	logger  *logger.Logger
}

// NewHandler creates a new health handler
func NewHandler(service health.HealthService) *Handler {
	return &Handler{
		service: service,
		logger:  logger.GetGlobal(),
	}
}

// Ensure Handler implements genhttp.ServerInterface
var _ genhttp.ServerInterface = (*Handler)(nil)

// GetHealth handles GET /health requests (implements genhttp.ServerInterface)
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get request ID for logging
	requestID := reqctx.GetRequestID(ctx)

	// Perform health check
	healthResponse := h.service.GetHealth(ctx)

	// Log the health check
	duration := time.Since(start)
	h.logger.InfoWithContext(ctx, "Health check completed",
		"requestId", requestID,
		"status", healthResponse.Status,
		"duration", duration.String(),
		"checksCount", len(healthResponse.Checks),
	)

	// Set appropriate HTTP status code
	var statusCode int
	switch healthResponse.Status {
	case health.StatusHealthy:
		statusCode = http.StatusOK
	case health.StatusDegraded:
		statusCode = http.StatusOK // Still OK but with warnings
	case health.StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		h.logger.ErrorWithContext(ctx, "Failed to encode health response",
			"requestId", requestID,
			"error", err.Error(),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetHealthLive handles GET /health/live requests (implements genhttp.ServerInterface)
func (h *Handler) GetHealthLive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := reqctx.GetRequestID(ctx)

	// Simple liveness check - just return OK if the service is running
	h.logger.DebugWithContext(ctx, "Liveness check",
		"requestId", requestID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"service":   "social-media-service",
	}

	json.NewEncoder(w).Encode(response)
}

// GetHealthReady handles GET /health/ready requests (implements genhttp.ServerInterface)
func (h *Handler) GetHealthReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := reqctx.GetRequestID(ctx)

	// Check if the service is ready to accept traffic
	healthResponse := h.service.GetHealth(ctx)

	var statusCode int
	var status string

	if healthResponse.Status == health.StatusHealthy {
		statusCode = http.StatusOK
		status = "ready"
	} else {
		statusCode = http.StatusServiceUnavailable
		status = "not ready"
	}

	h.logger.DebugWithContext(ctx, "Readiness check",
		"requestId", requestID,
		"status", status,
		"healthStatus", healthResponse.Status,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"service":   "social-media-service",
		"health":    healthResponse.Status,
	}

	json.NewEncoder(w).Encode(response)
}
