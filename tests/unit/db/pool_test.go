package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yourusername/air-go/internal/db"
)

// TestPoolConfiguration_MinMaxSize verifies pool size constraints (T095)
func TestPoolConfiguration_MinMaxSize(t *testing.T) {
	tests := []struct {
		name        string
		minPoolSize uint64
		maxPoolSize uint64
		shouldError bool
		description string
	}{
		{
			name:        "valid_range_min_5_max_10",
			minPoolSize: 5,
			maxPoolSize: 10,
			shouldError: false,
			description: "5-10 connections should be valid",
		},
		{
			name:        "valid_range_min_10_max_20",
			minPoolSize: 10,
			maxPoolSize: 20,
			shouldError: false,
			description: "10-20 connections should be valid (max allowed)",
		},
		{
			name:        "invalid_max_exceeds_20",
			minPoolSize: 5,
			maxPoolSize: 25,
			shouldError: true,
			description: "MaxPoolSize > 20 should fail validation",
		},
		{
			name:        "invalid_min_exceeds_max",
			minPoolSize: 15,
			maxPoolSize: 10,
			shouldError: true,
			description: "MinPoolSize > MaxPoolSize should fail validation",
		},
		{
			name:        "invalid_max_zero",
			minPoolSize: 5,
			maxPoolSize: 0,
			shouldError: true,
			description: "MaxPoolSize = 0 should fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      tt.minPoolSize,
				MaxPoolSize:      tt.maxPoolSize,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			}

			err := config.Validate()

			if tt.shouldError {
				assert.Error(t, err, "%s: validation should fail", tt.description)
			} else {
				assert.NoError(t, err, "%s: validation should pass", tt.description)
			}
		})
	}
}

// TestPoolConfiguration_IdleTimeout verifies idle connection timeout (T095)
func TestPoolConfiguration_IdleTimeout(t *testing.T) {
	tests := []struct {
		name            string
		maxConnIdleTime time.Duration
		shouldError     bool
		description     string
	}{
		{
			name:            "valid_5_minutes",
			maxConnIdleTime: 5 * time.Minute,
			shouldError:     false,
			description:     "5 minute idle timeout should be valid",
		},
		{
			name:            "valid_1_minute",
			maxConnIdleTime: 1 * time.Minute,
			shouldError:     false,
			description:     "1 minute idle timeout should be valid",
		},
		{
			name:            "valid_10_minutes",
			maxConnIdleTime: 10 * time.Minute,
			shouldError:     false,
			description:     "10 minute idle timeout should be valid",
		},
		{
			name:            "zero_timeout_valid",
			maxConnIdleTime: 0,
			shouldError:     false,
			description:     "Zero timeout (no idle timeout) should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  tt.maxConnIdleTime,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			}

			err := config.Validate()

			if tt.shouldError {
				assert.Error(t, err, "%s: validation should fail", tt.description)
			} else {
				assert.NoError(t, err, "%s: validation should pass", tt.description)
			}
		})
	}
}

// TestPoolConfiguration_ConnectionTimeouts verifies timeout settings (T095)
func TestPoolConfiguration_ConnectionTimeouts(t *testing.T) {
	tests := []struct {
		name             string
		connectTimeout   time.Duration
		operationTimeout time.Duration
		shouldError      bool
		description      string
	}{
		{
			name:             "valid_30s_connect_10s_operation",
			connectTimeout:   30 * time.Second,
			operationTimeout: 10 * time.Second,
			shouldError:      false,
			description:      "30s connect, 10s operation should be valid",
		},
		{
			name:             "valid_minimum_10s_5s",
			connectTimeout:   10 * time.Second,
			operationTimeout: 5 * time.Second,
			shouldError:      false,
			description:      "10s connect (minimum), 5s operation (minimum) should be valid",
		},
		{
			name:             "invalid_connect_too_short",
			connectTimeout:   5 * time.Second,
			operationTimeout: 10 * time.Second,
			shouldError:      true,
			description:      "Connect timeout < 10s should fail",
		},
		{
			name:             "invalid_connect_too_long",
			connectTimeout:   70 * time.Second,
			operationTimeout: 10 * time.Second,
			shouldError:      true,
			description:      "Connect timeout > 60s should fail",
		},
		{
			name:             "invalid_operation_too_short",
			connectTimeout:   30 * time.Second,
			operationTimeout: 500 * time.Millisecond,
			shouldError:      true,
			description:      "Operation timeout < 1s should fail",
		},
		{
			name:             "invalid_operation_too_long",
			connectTimeout:   30 * time.Second,
			operationTimeout: 35 * time.Second,
			shouldError:      true,
			description:      "Operation timeout > 30s should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   tt.connectTimeout,
				OperationTimeout: tt.operationTimeout,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			}

			err := config.Validate()

			if tt.shouldError {
				assert.Error(t, err, "%s: validation should fail", tt.description)
			} else {
				assert.NoError(t, err, "%s: validation should pass", tt.description)
			}
		})
	}
}

// TestPoolConfiguration_DefaultValues verifies default configuration (T095)
func TestPoolConfiguration_DefaultValues(t *testing.T) {
	// Test that default config values fall within acceptable ranges
	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "testdb",
		ConnectTimeout:   30 * time.Second, // Default connect timeout
		OperationTimeout: 10 * time.Second, // Default operation timeout
		MinPoolSize:      5,                // Default min pool size
		MaxPoolSize:      10,               // Default max pool size
		MaxConnIdleTime:  5 * time.Minute,  // Default idle timeout
		MaxRetryAttempts: 3,                // Default max retry attempts
		RetryBaseDelay:   1 * time.Second,  // Default base delay
		RetryMaxDelay:    10 * time.Second, // Default max delay
	}

	err := config.Validate()
	assert.NoError(t, err, "Default configuration should be valid")

	// Verify pool size constraints (FR-006)
	assert.GreaterOrEqual(t, config.MaxPoolSize, uint64(10),
		"Default max pool size should be >= 10")
	assert.LessOrEqual(t, config.MaxPoolSize, uint64(20),
		"Default max pool size should be <= 20")

	// Verify timeout constraints (FR-007, FR-018)
	assert.GreaterOrEqual(t, config.OperationTimeout, 5*time.Second,
		"Default operation timeout should be >= 5s")
	assert.LessOrEqual(t, config.OperationTimeout, 10*time.Second,
		"Default operation timeout should be <= 10s")

	assert.GreaterOrEqual(t, config.ConnectTimeout, 10*time.Second,
		"Default connect timeout should be >= 10s")
	assert.LessOrEqual(t, config.ConnectTimeout, 60*time.Second,
		"Default connect timeout should be <= 60s")
}

// TestPoolConfiguration_RecommendedProduction verifies production settings (T095)
func TestPoolConfiguration_RecommendedProduction(t *testing.T) {
	// Test recommended production configuration
	productionConfig := &db.DBConfig{
		URI:              "mongodb://prod-host:27017",
		Database:         "production",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      10, // Higher min for production
		MaxPoolSize:      20, // Max allowed
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	err := productionConfig.Validate()
	assert.NoError(t, err, "Production configuration should be valid")

	// Verify production settings are optimal
	assert.Equal(t, uint64(20), productionConfig.MaxPoolSize,
		"Production should use maximum allowed pool size")
	assert.GreaterOrEqual(t, productionConfig.MinPoolSize, uint64(10),
		"Production should maintain higher minimum pool size")
}
