package health

import "time"

// HealthStatus represents the health status of the application
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Service   string        `json:"service"`
	Status    HealthStatus  `json:"status"`
	Message   string        `json:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    HealthStatus  `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Version   string        `json:"version"`
	Uptime    time.Duration `json:"uptime"`
	Checks    []HealthCheck `json:"checks"`
}

// HealthService defines the interface for health operations
type HealthService interface {
	GetHealth() HealthResponse
	CheckDatabase() HealthCheck
	CheckRedis() HealthCheck
	CheckExternalAPI() HealthCheck
}

// HealthRepository defines the interface for health data operations
type HealthRepository interface {
	PingDatabase() error
	PingRedis() error
	PingExternalAPI() error
}
