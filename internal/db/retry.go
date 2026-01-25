package db

import (
	"math/rand"
	"time"

	"github.com/rs/zerolog"
)

// RetryState tracks retry attempts during connection failures
type RetryState struct {
	Attempt       int           // Current attempt number (1-3)
	LastError     error         // Most recent error
	NextRetryAt   time.Time     // When next retry will occur
	TotalDuration time.Duration // Cumulative retry time
}

// CalculateDelay calculates the retry delay with exponential backoff and jitter
// Retry schedule: 1s, 2s, 10s (with ±20% jitter)
// Exported for testing (T092)
func CalculateDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	var delay time.Duration

	switch attempt {
	case 1:
		delay = 1 * time.Second
	case 2:
		delay = 2 * time.Second
	case 3:
		delay = 10 * time.Second
	default:
		delay = maxDelay
	}

	// Add ±20% jitter to prevent thundering herd
	jitter := time.Duration(rand.Int63n(int64(delay * 40 / 100))) // 40% range
	jitter = jitter - (delay * 20 / 100)                          // Center at ±20%

	jittered := delay + jitter

	// Ensure we don't exceed max delay
	if jittered > maxDelay {
		return maxDelay
	}

	// Ensure minimum delay of 100ms
	if jittered < 100*time.Millisecond {
		return 100 * time.Millisecond
	}

	return jittered
}

// ShouldRetry determines if another retry attempt should be made
// Exported for testing (T092)
func ShouldRetry(attempt int, maxAttempts int) bool {
	return attempt < maxAttempts
}

// LogRetryAttempt logs a retry attempt with structured fields
// Exported for testing (T092)
func LogRetryAttempt(logger zerolog.Logger, state *RetryState, delay time.Duration) {
	logger.Warn().
		Int("attempt", state.Attempt).
		Err(state.LastError).
		Dur("next_retry_delay_ms", delay).
		Dur("cumulative_duration_ms", state.TotalDuration).
		Msg("Connection failed, will retry")
}
