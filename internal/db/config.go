package db

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBConfig holds MongoDB database configuration
type DBConfig struct {
	// Connection
	URI      string // MongoDB connection string
	Database string // Target database name

	// Timeouts (from FR-007)
	ConnectTimeout   time.Duration // Initial connection establishment (30s per spec)
	OperationTimeout time.Duration // Individual operation timeout (5-10s per spec)

	// Connection Pool (from FR-014)
	MinPoolSize     uint64        // Minimum connections (5 per research)
	MaxPoolSize     uint64        // Maximum connections (10-20 per spec)
	MaxConnIdleTime time.Duration // Connection idle timeout (5m per research)

	// Retry Configuration (from FR-010)
	MaxRetryAttempts int           // Maximum reconnection attempts (3 per spec)
	RetryBaseDelay   time.Duration // Initial retry delay (1s per research)
	RetryMaxDelay    time.Duration // Maximum retry delay (10s per research)
}

// Validate validates the entire configuration
func (c *DBConfig) Validate() error {
	if c == nil {
		return errors.New("configuration cannot be nil")
	}

	if err := validateURI(c.URI); err != nil {
		return fmt.Errorf("invalid URI: %w", err)
	}

	if err := validateDatabaseName(c.Database); err != nil {
		return fmt.Errorf("invalid database name: %w", err)
	}

	if err := validateTimeouts(c); err != nil {
		return err
	}

	if err := validatePoolSize(c); err != nil {
		return err
	}

	return nil
}

// validateURI validates the MongoDB connection string
func validateURI(uri string) error {
	if uri == "" {
		return errors.New("URI cannot be empty")
	}

	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		return errors.New("URI must start with mongodb:// or mongodb+srv://")
	}

	// Basic validation of URI format
	// Detailed validation will be performed by MongoDB driver during connection
	_ = options.Client().ApplyURI(uri)

	return nil
}

// validateDatabaseName validates the database name
func validateDatabaseName(name string) error {
	if name == "" {
		return errors.New("database name cannot be empty")
	}

	if len(name) > 64 {
		return errors.New("database name too long (max 64 characters)")
	}

	// Pattern: alphanumeric + underscore, must start with letter
	pattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !pattern.MatchString(name) {
		return errors.New("database name must start with letter and contain only alphanumeric + underscore")
	}

	// Reserved names
	reserved := []string{"admin", "local", "config"}
	for _, r := range reserved {
		if name == r {
			return fmt.Errorf("database name '%s' is reserved", name)
		}
	}

	return nil
}

// validateTimeouts validates timeout configurations
func validateTimeouts(config *DBConfig) error {
	// Connect timeout: 10s-60s (spec: 30s)
	if config.ConnectTimeout < 10*time.Second || config.ConnectTimeout > 60*time.Second {
		return fmt.Errorf("connect timeout must be between 10s and 60s, got %v", config.ConnectTimeout)
	}

	// Operation timeout: 1s-30s (spec: 5-10s)
	if config.OperationTimeout < 1*time.Second || config.OperationTimeout > 30*time.Second {
		return fmt.Errorf("operation timeout must be between 1s and 30s, got %v", config.OperationTimeout)
	}

	return nil
}

// validatePoolSize validates connection pool size configuration
func validatePoolSize(config *DBConfig) error {
	// Min pool size
	if config.MinPoolSize < 1 {
		return errors.New("min pool size must be at least 1")
	}

	// Max pool size: 10-20 per spec
	if config.MaxPoolSize < 10 || config.MaxPoolSize > 20 {
		return fmt.Errorf("max pool size must be between 10 and 20, got %d", config.MaxPoolSize)
	}

	// Min < Max
	if config.MinPoolSize >= config.MaxPoolSize {
		return fmt.Errorf("min pool size (%d) must be less than max pool size (%d)",
			config.MinPoolSize, config.MaxPoolSize)
	}

	return nil
}
