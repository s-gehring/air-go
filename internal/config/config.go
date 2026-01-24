package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Port        int
	LogFormat   string
	SchemaPath  string
	JWTSecret   string
	CORSOrigins []string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("SCHEMA_PATH", "./schema.graphqls")
	viper.SetDefault("CORS_ORIGINS", []string{"*"})

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
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
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
