package resolvers

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// Performance thresholds for different query types
const (
	SlowQueryThresholdSimple = 500 * time.Millisecond  // Simple lookups (by ID)
	SlowQueryThresholdSearch = 2000 * time.Millisecond // Search queries with filters
	SlowQueryThresholdHealth = 100 * time.Millisecond  // Health check queries
)

// getQueryThreshold returns the appropriate performance threshold for a query
func getQueryThreshold(queryName string) time.Duration {
	// Health and metadata queries
	if queryName == "alive" || queryName == "health" {
		return SlowQueryThresholdHealth
	}

	// Search queries (paginated queries with filters)
	if isSearchQuery(queryName) {
		return SlowQueryThresholdSearch
	}

	// Default to simple lookup threshold
	return SlowQueryThresholdSimple
}

// isSearchQuery determines if a query is a search/filter query
func isSearchQuery(queryName string) bool {
	searchQueries := map[string]bool{
		"referencePortfolioSearch":           true,
		"customerSearch":                     true,
		"employeeSearch":                     true,
		"employeeAllWithRoleGet":             true,
		"employeeAllByTeamleadGet":           true,
		"employeeAllByTeamleadAndTeamGet":    true,
		"employeeTeamMembersForTeamGet":      true,
		"teamSearch":                         true,
		"search":                             true, // inventory search
		"executionPlanSearch":                true,
		"openBankingTransactionsGet":         true,
		"customerOpenBankingProcessedDataGet": true,
	}
	return searchQueries[queryName]
}

// logQueryExecution logs query performance metrics
func logQueryExecution(ctx context.Context, queryName string, duration time.Duration, success bool) {
	threshold := getQueryThreshold(queryName)
	logEvent := log.Info()

	// Flag as slow query if exceeds threshold
	if duration > threshold {
		logEvent = log.Warn().Bool("slow_query", true)
	}

	// Extract request ID from context if available
	requestID := getRequestID(ctx)
	if requestID != "" {
		logEvent = logEvent.Str("request_id", requestID)
	}

	// Extract user ID from context if available
	if claims := getUserClaims(ctx); claims != nil {
		logEvent = logEvent.Str("user_id", claims.UserID)
	}

	logEvent.
		Str("query", queryName).
		Dur("duration_ms", duration).
		Dur("threshold_ms", threshold).
		Bool("success", success).
		Msg("GraphQL query executed")
}

// logQueryError logs query execution errors
func logQueryError(ctx context.Context, queryName string, err error, duration time.Duration) {
	logEvent := log.Error().Err(err)

	// Extract request ID from context if available
	requestID := getRequestID(ctx)
	if requestID != "" {
		logEvent = logEvent.Str("request_id", requestID)
	}

	// Extract user ID from context if available
	if claims := getUserClaims(ctx); claims != nil {
		logEvent = logEvent.Str("user_id", claims.UserID)
	}

	// Extract error code if it's a QueryError
	var queryErr *QueryError
	if err != nil {
		if qe, ok := err.(*QueryError); ok {
			queryErr = qe
			logEvent = logEvent.Str("error_code", qe.Code)
		}
	}

	logEvent.
		Str("query", queryName).
		Dur("duration_ms", duration).
		Bool("success", false).
		Msg("GraphQL query error")
}

// getRequestID extracts the request ID from context
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}
	return ""
}
