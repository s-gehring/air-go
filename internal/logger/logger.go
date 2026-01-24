package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup initializes the global logger with the specified format
func Setup(format string) {
	var output io.Writer = os.Stdout

	if format == "console" {
		// Human-readable console output for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// WithRequestID returns a logger with the request ID in context
func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}

// WithUserID returns a logger with the user ID in context
func WithUserID(logger zerolog.Logger, userID string) zerolog.Logger {
	return logger.With().Str("user_id", userID).Logger()
}

// WithOperation returns a logger with the operation name in context
func WithOperation(logger zerolog.Logger, operation string) zerolog.Logger {
	return logger.With().Str("operation", operation).Logger()
}

// WithDuration returns a logger with the duration in context
func WithDuration(logger zerolog.Logger, duration time.Duration) zerolog.Logger {
	return logger.With().Dur("duration_ms", duration).Logger()
}
