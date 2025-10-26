package config

import (
	"github.com/fanzru/social-media-service-go/pkg/env"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Storage  StorageConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	DBName             string
	SSLMode            string
	LogQueries         bool
	LogSlowQueries     bool
	SlowQueryThreshold int // in milliseconds
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration int // in hours
}

// StorageConfig holds file storage configuration
type StorageConfig struct {
	UploadPath  string
	MaxSize     int64 // in bytes
	AllowedExts []string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: env.GetInt("SERVER_PORT", 8080),
			Host: env.GetString("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:               env.GetString("DB_HOST", "localhost"),
			Port:               env.GetInt("DB_PORT", 5432),
			User:               env.GetString("DB_USER", "postgres"),
			Password:           env.GetString("DB_PASSWORD", "password"),
			DBName:             env.GetString("DB_NAME", "social_media"),
			SSLMode:            env.GetString("DB_SSL_MODE", "disable"),
			LogQueries:         env.GetBool("DB_LOG_QUERIES", true),
			LogSlowQueries:     env.GetBool("DB_LOG_SLOW_QUERIES", true),
			SlowQueryThreshold: env.GetInt("DB_SLOW_QUERY_THRESHOLD", 100), // 100ms default
		},
		JWT: JWTConfig{
			Secret:     env.GetString("JWT_SECRET", "your-secret-key"),
			Expiration: env.GetInt("JWT_EXPIRATION", 24),
		},
		Storage: StorageConfig{
			UploadPath:  env.GetString("UPLOAD_PATH", "./uploads"),
			MaxSize:     env.GetInt64("MAX_FILE_SIZE", 104857600), // 100MB
			AllowedExts: env.GetStringSlice("ALLOWED_EXTENSIONS", []string{".png", ".jpg", ".jpeg", ".bmp"}),
		},
	}
}
