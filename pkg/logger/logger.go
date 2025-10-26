package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/reqctx"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  LogLevel
	Output io.Writer
	Format string // "json" or "text"
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  LevelInfo,
		Output: os.Stdout,
		Format: "json",
	}
}

// New creates a new logger instance
func New(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	var handler slog.Handler

	// Create JSON handler with custom options
	opts := &slog.HandlerOptions{
		Level:     parseLevel(config.Level),
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format to RFC 3339
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			// Customize level format
			if a.Key == slog.LevelKey {
				return slog.Attr{
					Key:   "level",
					Value: slog.StringValue(a.Value.String()),
				}
			}
			// Customize message format
			if a.Key == slog.MessageKey {
				return slog.Attr{
					Key:   "message",
					Value: a.Value,
				}
			}
			return a
		},
	}

	if config.Format == "json" {
		handler = slog.NewJSONHandler(config.Output, opts)
	} else {
		handler = slog.NewTextHandler(config.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewFromEnv creates a logger from environment variables
func NewFromEnv() *Logger {
	config := DefaultConfig()

	// Parse log level from environment
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevel(level)
	}

	// Parse log format from environment
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = format
	}

	return New(config)
}

// WithRequestID adds request ID to the logger context
func (l *Logger) WithRequestID(ctx context.Context) *slog.Logger {
	requestID := reqctx.GetRequestID(ctx)
	if requestID != "" {
		return l.Logger.With("requestId", requestID)
	}
	return l.Logger
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *slog.Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return l.Logger.With(args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
}

// DebugWithContext logs a debug message with request context
func (l *Logger) DebugWithContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithRequestID(ctx).Debug(msg, args...)
}

// InfoWithContext logs an info message with request context
func (l *Logger) InfoWithContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithRequestID(ctx).Info(msg, args...)
}

// WarnWithContext logs a warning message with request context
func (l *Logger) WarnWithContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithRequestID(ctx).Warn(msg, args...)
}

// ErrorWithContext logs an error message with request context
func (l *Logger) ErrorWithContext(ctx context.Context, msg string, args ...interface{}) {
	l.WithRequestID(ctx).Error(msg, args...)
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	l.WithRequestID(ctx).Info("HTTP Request",
		"method", method,
		"path", path,
		"statusCode", statusCode,
		"duration", duration.String(),
	)
}

// LogError logs error with context
func (l *Logger) LogError(ctx context.Context, err error, msg string, fields ...map[string]interface{}) {
	args := []interface{}{"error", err.Error()}

	// Add additional fields if provided
	if len(fields) > 0 {
		for k, v := range fields[0] {
			args = append(args, k, v)
		}
	}

	l.WithRequestID(ctx).Error(msg, args...)
}

// LogDatabase logs database operation
func (l *Logger) LogDatabase(ctx context.Context, operation string, table string, duration time.Duration, err error) {
	args := []interface{}{
		"operation", operation,
		"table", table,
		"duration", duration.String(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		l.WithRequestID(ctx).Error("Database operation failed", args...)
	} else {
		l.WithRequestID(ctx).Debug("Database operation completed", args...)
	}
}

// LogService logs service operation
func (l *Logger) LogService(ctx context.Context, service, method string, duration time.Duration, err error) {
	args := []interface{}{
		"service", service,
		"method", method,
		"duration", duration.String(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		l.WithRequestID(ctx).Error("Service operation failed", args...)
	} else {
		l.WithRequestID(ctx).Debug("Service operation completed", args...)
	}
}

// parseLevel converts LogLevel to slog.Level
func parseLevel(level LogLevel) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger
func Init(config *Config) {
	globalLogger = New(config)
}

// InitFromEnv initializes the global logger from environment
func InitFromEnv() {
	globalLogger = NewFromEnv()
}

// GetGlobal returns the global logger instance
func GetGlobal() *Logger {
	if globalLogger == nil {
		globalLogger = NewFromEnv()
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(msg string, args ...interface{}) {
	GetGlobal().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetGlobal().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetGlobal().Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetGlobal().Error(msg, args...)
}

func DebugWithContext(ctx context.Context, msg string, args ...interface{}) {
	GetGlobal().DebugWithContext(ctx, msg, args...)
}

func InfoWithContext(ctx context.Context, msg string, args ...interface{}) {
	GetGlobal().InfoWithContext(ctx, msg, args...)
}

func WarnWithContext(ctx context.Context, msg string, args ...interface{}) {
	GetGlobal().WarnWithContext(ctx, msg, args...)
}

func ErrorWithContext(ctx context.Context, msg string, args ...interface{}) {
	GetGlobal().ErrorWithContext(ctx, msg, args...)
}

func LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	GetGlobal().LogRequest(ctx, method, path, statusCode, duration)
}

func LogError(ctx context.Context, err error, msg string, fields ...map[string]interface{}) {
	GetGlobal().LogError(ctx, err, msg, fields...)
}

func LogDatabase(ctx context.Context, operation string, table string, duration time.Duration, err error) {
	GetGlobal().LogDatabase(ctx, operation, table, duration, err)
}

func LogService(ctx context.Context, service, method string, duration time.Duration, err error) {
	GetGlobal().LogService(ctx, service, method, duration, err)
}
