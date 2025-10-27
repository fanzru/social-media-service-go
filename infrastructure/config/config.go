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
	StatsD   StatsDConfig
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
	MaxSize     int64 // in bytes
	AllowedExts []string

	// S3 Configuration
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Endpoint        string
	S3ImageBaseURL    string

	// Image Processing Configuration
	ImageResizeWidth  int
	ImageResizeHeight int
	ImageQuality      int
}

// StatsDConfig holds StatsD configuration
type StatsDConfig struct {
	Host     string
	Port     int
	Prefix   string
	Sampling float64
	Enabled  bool
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
			MaxSize:     env.GetInt64("MAX_FILE_SIZE", 104857600), // 100MB
			AllowedExts: env.GetStringSlice("ALLOWED_EXTENSIONS", []string{".png", ".jpg", ".jpeg", ".bmp"}),

			// S3 Configuration
			S3Region:          env.GetString("S3_REGION", "auto"),
			S3Bucket:          env.GetString("S3_BUCKET", "social-media"),
			S3AccessKeyID:     env.GetString("S3_ACCESS_KEY_ID", ""),
			S3SecretAccessKey: env.GetString("S3_SECRET_ACCESS_KEY", ""),
			S3Endpoint:        env.GetString("S3_ENDPOINT", ""),
			S3ImageBaseURL:    env.GetString("S3_IMAGE_BASE_URL", ""),

			// Image Processing Configuration
			ImageResizeWidth:  env.GetInt("IMAGE_RESIZE_WIDTH", 600),
			ImageResizeHeight: env.GetInt("IMAGE_RESIZE_HEIGHT", 600),
			ImageQuality:      env.GetInt("IMAGE_QUALITY", 85),
		},
		StatsD: StatsDConfig{
			Host:     env.GetString("STATSD_HOST", "localhost"),
			Port:     env.GetInt("STATSD_PORT", 8125),
			Prefix:   env.GetString("STATSD_PREFIX", "social_media"),
			Sampling: env.GetFloat64("STATSD_SAMPLING", 1.0),
			Enabled:  env.GetBool("STATSD_ENABLED", true),
		},
	}
}
