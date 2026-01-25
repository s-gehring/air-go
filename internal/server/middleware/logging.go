package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	// RequestIDKey is the context key for storing the request ID
	RequestIDKey ContextKey = "request_id"
	// RequestIDHeader is the HTTP header name for the request ID
	RequestIDHeader = "X-Request-ID"
)

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// LoggingMiddleware logs HTTP requests and responses with request ID, duration, and other metadata
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		w.Header().Set(RequestIDHeader, requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Wrap the response writer to capture status code and size
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200 if WriteHeader is not called
			bytesWritten:   0,
		}

		// Record start time
		startTime := time.Now()

		// Log incoming request
		log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("Incoming request")

		// Call the next handler
		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

		// Calculate duration
		duration := time.Since(startTime)

		// Log response
		logEvent := log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrappedWriter.statusCode).
			Int("bytes", wrappedWriter.bytesWritten).
			Dur("duration_ms", duration)

		// Add user ID if available
		if userID, ok := ctx.Value(UserIDKey).(string); ok {
			logEvent = logEvent.Str("user_id", userID)
		}

		logEvent.Msg("Request completed")
	})
}

// GetRequestID extracts the request ID from the request context
func GetRequestID(r *http.Request) string {
	requestID, ok := r.Context().Value(RequestIDKey).(string)
	if !ok {
		return ""
	}
	return requestID
}
