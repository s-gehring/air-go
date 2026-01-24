package db_test

import (
	"testing"
	"time"

	"github.com/yourusername/air-go/internal/db"
)

// TestDBConfig_Validate_ValidConfiguration tests valid configuration scenarios
func TestDBConfig_Validate_ValidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config *db.DBConfig
	}{
		{
			name: "minimal valid configuration",
			config: &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			},
		},
		{
			name: "maximum pool size (20)",
			config: &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      20,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			},
		},
		{
			name: "longer timeouts",
			config: &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   60 * time.Second,
				OperationTimeout: 30 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      15,
				MaxConnIdleTime:  10 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			},
		},
		{
			name: "mongodb+srv URI scheme",
			config: &db.DBConfig{
				URI:              "mongodb+srv://user:pass@cluster.mongodb.net/mydb",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v, expected nil", err)
			}
		})
	}
}

// TestDBConfig_Validate_InvalidURI tests invalid URI scenarios
func TestDBConfig_Validate_InvalidURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{
			name: "empty URI",
			uri:  "",
		},
		{
			name: "invalid scheme",
			uri:  "http://localhost:27017",
		},
		{
			name: "missing scheme",
			uri:  "localhost:27017",
		},
		{
			name: "invalid format",
			uri:  "not a uri at all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &db.DBConfig{
				URI:              tt.uri,
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			}

			err := config.Validate()
			if err == nil {
				t.Error("Validate() expected error for invalid URI, got nil")
			}
		})
	}
}

// TestDBConfig_Validate_InvalidDatabase tests invalid database name scenarios
func TestDBConfig_Validate_InvalidDatabase(t *testing.T) {
	tests := []struct {
		name     string
		database string
	}{
		{
			name:     "empty database name",
			database: "",
		},
		{
			name:     "database name with spaces",
			database: "my database",
		},
		{
			name:     "database name with special chars",
			database: "my$database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         tt.database,
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			}

			err := config.Validate()
			if err == nil {
				t.Errorf("Validate() expected error for invalid database name %q, got nil", tt.database)
			}
		})
	}
}

// TestDBConfig_Validate_InvalidTimeouts tests invalid timeout scenarios
func TestDBConfig_Validate_InvalidTimeouts(t *testing.T) {
	tests := []struct {
		name             string
		connectTimeout   time.Duration
		operationTimeout time.Duration
	}{
		{
			name:             "zero connect timeout",
			connectTimeout:   0,
			operationTimeout: 10 * time.Second,
		},
		{
			name:             "negative connect timeout",
			connectTimeout:   -1 * time.Second,
			operationTimeout: 10 * time.Second,
		},
		{
			name:             "zero operation timeout",
			connectTimeout:   30 * time.Second,
			operationTimeout: 0,
		},
		{
			name:             "negative operation timeout",
			connectTimeout:   30 * time.Second,
			operationTimeout: -1 * time.Second,
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
			if err == nil {
				t.Error("Validate() expected error for invalid timeouts, got nil")
			}
		})
	}
}

// TestDBConfig_Validate_InvalidPoolSize tests invalid pool size scenarios
func TestDBConfig_Validate_InvalidPoolSize(t *testing.T) {
	tests := []struct {
		name        string
		minPoolSize uint64
		maxPoolSize uint64
	}{
		{
			name:        "max pool size too large",
			minPoolSize: 5,
			maxPoolSize: 21,
		},
		{
			name:        "max pool size zero",
			minPoolSize: 5,
			maxPoolSize: 0,
		},
		{
			name:        "min larger than max",
			minPoolSize: 15,
			maxPoolSize: 10,
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
			if err == nil {
				t.Error("Validate() expected error for invalid pool sizes, got nil")
			}
		})
	}
}

// TestDBConfig_Validate_NilConfig tests nil configuration
func TestDBConfig_Validate_NilConfig(t *testing.T) {
	var config *db.DBConfig
	err := config.Validate()
	if err == nil {
		t.Error("Validate() expected error for nil config, got nil")
	}
}
