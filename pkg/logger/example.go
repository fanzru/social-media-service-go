package logger

import (
	"context"
	"time"
)

// Example usage of the logger package
func ExampleUsage() {
	// Initialize logger
	logger := New(DefaultConfig())
	
	// Basic logging
	logger.Info("Application started")
	logger.Debug("Debug information", "key", "value")
	logger.Warn("Warning message", "warning", "something happened")
	logger.Error("Error occurred", "error", "something went wrong")
	
	// Context-based logging
	ctx := context.Background()
	logger.InfoWithContext(ctx, "Request processed")
	
	// Request logging
	logger.LogRequest(ctx, "POST", "/api/account/register", 201, 150*time.Millisecond)
	
	// Database logging
	logger.LogDatabase(ctx, "INSERT", "accounts", 50*time.Millisecond, nil)
	logger.LogDatabase(ctx, "SELECT", "accounts", 30*time.Millisecond, nil)
	
	// Service logging
	logger.LogService(ctx, "AccountService", "Register", 100*time.Millisecond, nil)
	
	// Error logging with additional fields
	logger.LogError(ctx, nil, "Failed to process request", map[string]interface{}{
		"userId": 123,
		"action": "register",
	})
	
	// Custom fields
	logger.WithFields(map[string]interface{}{
		"userId":    123,
		"sessionId": "abc123",
		"action":    "login",
	}).Info("User action performed")
}

// Example of JSON output format:
/*
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "INFO",
  "message": "Application started",
  "requestId": "req-123456789"
}

{
  "timestamp": "2024-01-01T12:00:01Z",
  "level": "INFO",
  "message": "HTTP Request",
  "requestId": "req-123456789",
  "method": "POST",
  "path": "/api/account/register",
  "statusCode": 201,
  "duration": "150ms"
}

{
  "timestamp": "2024-01-01T12:00:02Z",
  "level": "DEBUG",
  "message": "Database operation completed",
  "requestId": "req-123456789",
  "operation": "INSERT",
  "table": "accounts",
  "duration": "50ms"
}

{
  "timestamp": "2024-01-01T12:00:03Z",
  "level": "ERROR",
  "message": "Service operation failed",
  "requestId": "req-123456789",
  "service": "AccountService",
  "method": "Register",
  "duration": "100ms",
  "error": "email already exists"
}
*/
