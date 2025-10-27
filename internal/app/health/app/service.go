package app

import (
	"context"
	"time"

	"github.com/fanzru/social-media-service-go/internal/app/health"
)

// Service implements health service interface
type Service struct {
	repo health.HealthRepository
}

// NewService creates a new health service
func NewService(repo health.HealthRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetHealth returns the overall health status
func (s *Service) GetHealth(ctx context.Context) health.HealthResponse {
	startTime := time.Now()

	// Perform health checks
	checks := []health.HealthCheck{
		s.CheckDatabase(ctx),
		s.CheckRedis(ctx),
		s.CheckExternalAPI(ctx),
	}

	// Determine overall status
	overallStatus := health.StatusHealthy
	for _, check := range checks {
		if check.Status == health.StatusUnhealthy {
			overallStatus = health.StatusUnhealthy
			break
		} else if check.Status == health.StatusDegraded {
			overallStatus = health.StatusDegraded
		}
	}

	return health.HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime),
		Checks:    checks,
	}
}

// CheckDatabase checks database connectivity
func (s *Service) CheckDatabase(ctx context.Context) health.HealthCheck {
	start := time.Now()
	err := s.repo.PingDatabase(ctx)
	duration := time.Since(start)

	check := health.HealthCheck{
		Service:   "database",
		Timestamp: time.Now(),
		Duration:  duration,
	}

	if err != nil {
		check.Status = health.StatusUnhealthy
		check.Message = err.Error()
	} else {
		check.Status = health.StatusHealthy
		check.Message = "Database connection is healthy"
	}

	return check
}

// CheckRedis checks Redis connectivity
func (s *Service) CheckRedis(ctx context.Context) health.HealthCheck {
	start := time.Now()
	err := s.repo.PingRedis(ctx)
	duration := time.Since(start)

	check := health.HealthCheck{
		Service:   "redis",
		Timestamp: time.Now(),
		Duration:  duration,
	}

	if err != nil {
		check.Status = health.StatusDegraded
		check.Message = err.Error()
	} else {
		check.Status = health.StatusHealthy
		check.Message = "Redis connection is healthy"
	}

	return check
}

// CheckExternalAPI checks external API connectivity
func (s *Service) CheckExternalAPI(ctx context.Context) health.HealthCheck {
	start := time.Now()
	err := s.repo.PingExternalAPI(ctx)
	duration := time.Since(start)

	check := health.HealthCheck{
		Service:   "external-api",
		Timestamp: time.Now(),
		Duration:  duration,
	}

	if err != nil {
		check.Status = health.StatusDegraded
		check.Message = err.Error()
	} else {
		check.Status = health.StatusHealthy
		check.Message = "External API connection is healthy"
	}

	return check
}
