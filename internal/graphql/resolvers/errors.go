package resolvers

import (
	"errors"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/mongo"
)

// Error codes for GraphQL responses
const (
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeDatabaseError       = "DATABASE_ERROR"
	ErrCodeExternalService     = "EXTERNAL_SERVICE_ERROR"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
)

// QueryError represents a custom GraphQL error with an error code
type QueryError struct {
	Message string
	Code    string
	Cause   error
}

// Error implements the error interface
func (e *QueryError) Error() string {
	return e.Message
}

// Unwrap allows errors.Is and errors.As to work with QueryError
func (e *QueryError) Unwrap() error {
	return e.Cause
}

// Extensions returns the error extensions for GraphQL response
func (e *QueryError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code": e.Code,
	}
}

// mapMongoError maps MongoDB errors to GraphQL errors with appropriate error codes
func mapMongoError(err error) error {
	if err == nil {
		return nil
	}

	// Handle MongoDB no documents error (entity not found)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &QueryError{
			Message: "Entity not found",
			Code:    ErrCodeNotFound,
			Cause:   err,
		}
	}

	// Handle MongoDB connection errors
	if mongo.IsTimeout(err) || mongo.IsNetworkError(err) {
		return &QueryError{
			Message: "Database connection failed",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}

	// Default database error
	return &QueryError{
		Message: "Database operation failed",
		Code:    ErrCodeDatabaseError,
		Cause:   err,
	}
}

// newInvalidInputError creates a new invalid input error
func newInvalidInputError(message string) error {
	return &QueryError{
		Message: message,
		Code:    ErrCodeInvalidInput,
	}
}

// newUnauthorizedError creates a new unauthorized error
func newUnauthorizedError(message string) error {
	return &QueryError{
		Message: message,
		Code:    ErrCodeUnauthorized,
	}
}

// newForbiddenError creates a new forbidden error
func newForbiddenError(message string) error {
	return &QueryError{
		Message: message,
		Code:    ErrCodeForbidden,
	}
}

// newExternalServiceError creates a new external service error
func newExternalServiceError(message string, cause error) error {
	return &QueryError{
		Message: message,
		Code:    ErrCodeExternalService,
		Cause:   cause,
	}
}

// toGraphQLError converts a QueryError to gqlerror.Error
func toGraphQLError(err error) *gqlerror.Error {
	if err == nil {
		return nil
	}

	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		return &gqlerror.Error{
			Message: queryErr.Message,
			Extensions: map[string]interface{}{
				"code": queryErr.Code,
			},
		}
	}

	// Default error
	return &gqlerror.Error{
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"code": ErrCodeInternalServerError,
		},
	}
}
