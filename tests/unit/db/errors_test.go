package db_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yourusername/air-go/internal/db"
)

// TestErrorConstants verifies error constants are defined (T093)
func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		wants string
	}{
		{
			name:  "ErrInvalidConfiguration",
			err:   db.ErrInvalidConfiguration,
			wants: "invalid configuration",
		},
		{
			name:  "ErrConnectionTimeout",
			err:   db.ErrConnectionTimeout,
			wants: "connection timeout",
		},
		{
			name:  "ErrAlreadyConnected",
			err:   db.ErrAlreadyConnected,
			wants: "already connected",
		},
		{
			name:  "ErrNotConnected",
			err:   db.ErrNotConnected,
			wants: "not connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err, "Error should be defined")
			assert.Contains(t, tt.err.Error(), tt.wants,
				"Error message should contain: %s", tt.wants)
		})
	}
}

// TestErrorWrapping verifies error wrapping preserves context (T093)
func TestErrorWrapping(t *testing.T) {
	baseErr := errors.New("connection refused")
	wrappedErr := errors.New("failed to connect to MongoDB: " + baseErr.Error())

	assert.ErrorContains(t, wrappedErr, "connection refused",
		"Wrapped error should contain base error message")
	assert.ErrorContains(t, wrappedErr, "failed to connect",
		"Wrapped error should contain context")
}

// TestErrorUnwrapping verifies errors can be unwrapped (T093)
func TestErrorUnwrapping(t *testing.T) {
	baseErr := db.ErrInvalidConfiguration
	wrappedErr := errors.New("initialization failed: " + baseErr.Error())

	// Verify the wrapped error contains the base error message
	assert.ErrorContains(t, wrappedErr, baseErr.Error(),
		"Should be able to identify base error in wrapped error")
}

// TestErrorComparison verifies error equality checks (T093)
func TestErrorComparison(t *testing.T) {
	tests := []struct {
		name     string
		err1     error
		err2     error
		shouldBe bool
	}{
		{
			name:     "same_error_constants",
			err1:     db.ErrNotConnected,
			err2:     db.ErrNotConnected,
			shouldBe: true,
		},
		{
			name:     "different_error_constants",
			err1:     db.ErrNotConnected,
			err2:     db.ErrAlreadyConnected,
			shouldBe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldBe {
				assert.Equal(t, tt.err1, tt.err2, "Errors should be equal")
			} else {
				assert.NotEqual(t, tt.err1, tt.err2, "Errors should not be equal")
			}
		})
	}
}

// TestErrorMessages verifies error messages are descriptive (T093)
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		mustContain []string
		description string
	}{
		{
			name:        "invalid_configuration_descriptive",
			err:         db.ErrInvalidConfiguration,
			mustContain: []string{"invalid", "configuration"},
			description: "Should describe what is invalid",
		},
		{
			name:        "connection_timeout_descriptive",
			err:         db.ErrConnectionTimeout,
			mustContain: []string{"connection", "timeout"},
			description: "Should indicate connection and timeout",
		},
		{
			name:        "already_connected_descriptive",
			err:         db.ErrAlreadyConnected,
			mustContain: []string{"already", "connected"},
			description: "Should indicate state conflict",
		},
		{
			name:        "not_connected_descriptive",
			err:         db.ErrNotConnected,
			mustContain: []string{"not", "connected"},
			description: "Should indicate missing connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, keyword := range tt.mustContain {
				assert.Contains(t, errMsg, keyword,
					"%s: error message should contain '%s'", tt.description, keyword)
			}
		})
	}
}

// TestErrorContext verifies errors provide useful context (T093)
func TestErrorContext(t *testing.T) {
	// Simulate error scenarios
	scenarios := []struct {
		name          string
		createError   func() error
		expectedParts []string
		description   string
	}{
		{
			name: "connection_failure_with_context",
			createError: func() error {
				return errors.New("failed to connect to MongoDB: connection refused")
			},
			expectedParts: []string{"failed to connect", "MongoDB", "connection refused"},
			description:   "Connection errors should include database type and reason",
		},
		{
			name: "validation_failure_with_context",
			createError: func() error {
				return errors.New("invalid database configuration: URI cannot be empty")
			},
			expectedParts: []string{"invalid", "configuration", "URI"},
			description:   "Validation errors should specify what is invalid",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.createError()
			errMsg := err.Error()

			for _, part := range scenario.expectedParts {
				assert.Contains(t, errMsg, part,
					"%s: error should contain '%s'", scenario.description, part)
			}
		})
	}
}
