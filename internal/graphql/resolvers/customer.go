package resolvers

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/yourusername/air-go/internal/graphql/generated"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UUID validation regex pattern (RFC4122 format, case-insensitive)
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// isValidUUID validates UUID format according to RFC4122
// Supports uppercase, lowercase, and mixed case UUIDs
func isValidUUID(uuid string) bool {
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// customerGet retrieves a customer by identifier from MongoDB
// Returns nil for non-existent or deleted customers
// Returns error for invalid input or database failures
func customerGet(r *queryResolver, ctx context.Context, identifier string) (*generated.Customer, error) {
	// Start performance logging
	startTime := time.Now()
	var err error
	defer func() {
		duration := time.Since(startTime)
		logQueryExecution(ctx, "customerGet", duration, err == nil)
	}()

	// Validate UUID format (FR-005)
	if !isValidUUID(identifier) {
		err = newInvalidInputError("invalid UUID format")
		return nil, err
	}

	// Get customers collection
	collection := r.DBClient.Collection("customers")
	if collection == nil {
		err = &QueryError{
			Message: "Database not available",
			Code:    ErrCodeDatabaseError,
		}
		return nil, err
	}

	// Build query filter
	// - Match by identifier
	// - Exclude deleted customers (status.deletion = DELETED)
	filter := bson.M{
		"identifier":      identifier,
		"status.deletion": bson.M{"$ne": "DELETED"},
	}

	// Execute FindOne query
	result := collection.FindOne(ctx, filter)
	if result.Err() == mongo.ErrNoDocuments {
		// Customer not found or deleted - return nil (FR-004, FR-003)
		return nil, nil
	}
	if result.Err() != nil {
		err = mapMongoError(result.Err())
		return nil, err
	}

	// Decode customer document
	var customer generated.Customer
	if decodeErr := result.Decode(&customer); decodeErr != nil {
		err = mapMongoError(decodeErr)
		return nil, err
	}

	return &customer, nil
}
