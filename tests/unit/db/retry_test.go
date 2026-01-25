package db_test

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/yourusername/air-go/internal/db"
)

// TestCalculateDelay_RetrySchedule verifies retry schedule (1s, 2s, 10s) with jitter (T092)
func TestCalculateDelay_RetrySchedule(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 10 * time.Second

	tests := []struct {
		name        string
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
		description string
	}{
		{
			name:        "first_attempt",
			attempt:     1,
			expectedMin: 800 * time.Millisecond,  // 1s - 20% jitter
			expectedMax: 1200 * time.Millisecond, // 1s + 20% jitter
			description: "First attempt should use 1s delay with ±20% jitter",
		},
		{
			name:        "second_attempt",
			attempt:     2,
			expectedMin: 1600 * time.Millisecond, // 2s - 20% jitter
			expectedMax: 2400 * time.Millisecond, // 2s + 20% jitter
			description: "Second attempt should use 2s delay with ±20% jitter",
		},
		{
			name:        "third_attempt",
			attempt:     3,
			expectedMin: 8 * time.Second,  // 10s - 20% jitter
			expectedMax: 10 * time.Second, // 10s (capped at maxDelay)
			description: "Third attempt should use 10s delay with ±20% jitter",
		},
		{
			name:        "exceeds_max_attempts",
			attempt:     10,
			expectedMin: 8 * time.Second,  // maxDelay - 20% jitter
			expectedMax: 10 * time.Second, // maxDelay (capped)
			description: "Delay should be capped at maxDelay for attempts > 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to account for jitter randomness
			for i := 0; i < 10; i++ {
				delay := db.CalculateDelay(tt.attempt, baseDelay, maxDelay)

				assert.GreaterOrEqual(t, delay, tt.expectedMin,
					"%s: delay should be >= %v", tt.description, tt.expectedMin)
				assert.LessOrEqual(t, delay, tt.expectedMax,
					"%s: delay should be <= %v", tt.description, tt.expectedMax)
			}
		})
	}
}

// TestCalculateDelay_Jitter verifies jitter adds randomness (T092)
func TestCalculateDelay_Jitter(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 10 * time.Second
	attempt := 1

	// Collect multiple delay values to verify jitter creates variance
	delays := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		delay := db.CalculateDelay(attempt, baseDelay, maxDelay)
		delays[delay] = true
	}

	// With jitter, we should see multiple different delay values
	// (not just a single fixed value)
	assert.Greater(t, len(delays), 1,
		"Jitter should produce varying delay values, got only %d unique values", len(delays))
}

// TestShouldRetry_MaxAttempts verifies retry limit of 3 attempts (T092)
func TestShouldRetry_MaxAttempts(t *testing.T) {
	maxAttempts := 3

	tests := []struct {
		name     string
		attempt  int
		expected bool
	}{
		{
			name:     "attempt_1_should_retry",
			attempt:  1,
			expected: true,
		},
		{
			name:     "attempt_2_should_retry",
			attempt:  2,
			expected: true,
		},
		{
			name:     "attempt_3_should_not_retry",
			attempt:  3,
			expected: false,
		},
		{
			name:     "attempt_4_should_not_retry",
			attempt:  4,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.ShouldRetry(tt.attempt, maxAttempts)
			assert.Equal(t, tt.expected, result,
				"Attempt %d should retry: %v", tt.attempt, tt.expected)
		})
	}
}

// TestRetryState_Tracking verifies RetryState tracks attempt info (T092)
func TestRetryState_Tracking(t *testing.T) {
	state := &db.RetryState{
		Attempt:       2,
		LastError:     assert.AnError,
		TotalDuration: 5 * time.Second,
	}

	assert.Equal(t, 2, state.Attempt, "Should track current attempt")
	assert.Equal(t, assert.AnError, state.LastError, "Should track last error")
	assert.Equal(t, 5*time.Second, state.TotalDuration, "Should track total duration")
}

// TestLogRetryAttempt_StructuredLogging verifies retry logging (T092)
func TestLogRetryAttempt_StructuredLogging(t *testing.T) {
	// This test verifies that LogRetryAttempt can be called without panicking
	// Actual log output verification is done in logging_test.go

	logger := zerolog.Nop()

	state := &db.RetryState{
		Attempt:       2,
		LastError:     assert.AnError,
		TotalDuration: 3 * time.Second,
	}
	delay := 2 * time.Second

	// Should not panic
	assert.NotPanics(t, func() {
		db.LogRetryAttempt(logger, state, delay)
	}, "LogRetryAttempt should not panic")
}
