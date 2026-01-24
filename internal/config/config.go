package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/yourusername/air-go/internal/db"
)

// Config holds all configuration for the application
type Config struct {
	Port        int
	LogFormat   string
	SchemaPath  string
	JWTSecret   string
	CORSOrigins []string
	Database    *db.DBConfig // MongoDB configuration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("SCHEMA_PATH", "./schema.graphqls")
	viper.SetDefault("CORS_ORIGINS", []string{"*"})

	// MongoDB defaults
	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "air_dev")
	viper.SetDefault("MONGODB_TIMEOUT_CONNECT", "30s")
	viper.SetDefault("MONGODB_TIMEOUT_OPERATION", "10s")
	viper.SetDefault("MONGODB_POOL_MIN", 5)
	viper.SetDefault("MONGODB_POOL_MAX", 20)
	viper.SetDefault("MONGODB_POOL_IDLE_TIMEOUT", "5m")
	viper.SetDefault("MONGODB_RETRY_ATTEMPTS", 3)
	viper.SetDefault("MONGODB_RETRY_BASE_DELAY", "1s")
	viper.SetDefault("MONGODB_RETRY_MAX_DELAY", "10s")

	viper.AutomaticEnv()

	// Load from .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}
	}

	cfg := &Config{
		Port:        viper.GetInt("PORT"),
		LogFormat:   viper.GetString("LOG_FORMAT"),
		SchemaPath:  viper.GetString("SCHEMA_PATH"),
		JWTSecret:   viper.GetString("JWT_SECRET"),
		CORSOrigins: viper.GetStringSlice("CORS_ORIGINS"),
		Database: &db.DBConfig{
			URI:              viper.GetString("MONGODB_URI"),
			Database:         viper.GetString("MONGODB_DATABASE"),
			ConnectTimeout:   viper.GetDuration("MONGODB_TIMEOUT_CONNECT"),
			OperationTimeout: viper.GetDuration("MONGODB_TIMEOUT_OPERATION"),
			MinPoolSize:      uint64(viper.GetInt("MONGODB_POOL_MIN")),
			MaxPoolSize:      uint64(viper.GetInt("MONGODB_POOL_MAX")),
			MaxConnIdleTime:  viper.GetDuration("MONGODB_POOL_IDLE_TIMEOUT"),
			MaxRetryAttempts: viper.GetInt("MONGODB_RETRY_ATTEMPTS"),
			RetryBaseDelay:   viper.GetDuration("MONGODB_RETRY_BASE_DELAY"),
			RetryMaxDelay:    viper.GetDuration("MONGODB_RETRY_MAX_DELAY"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Validate database configuration
	if err := cfg.Database.Validate(); err != nil {
		return nil, fmt.Errorf("database configuration invalid: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Port < 1024 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1024 and 65535, got %d", c.Port)
	}

	if c.LogFormat != "json" && c.LogFormat != "console" {
		return fmt.Errorf("LOG_FORMAT must be 'json' or 'console', got '%s'", c.LogFormat)
	}

	if c.SchemaPath == "" {
		return fmt.Errorf("SCHEMA_PATH is required")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET should be at least 32 characters long for security, got %d characters", len(c.JWTSecret))
	}

	return nil
}
