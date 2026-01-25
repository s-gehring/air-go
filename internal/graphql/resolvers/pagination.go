package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PageInfo represents pagination metadata
type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     *string
	EndCursor       *string
}

// encodeCursor encodes an ObjectID to a base64 cursor string
func encodeCursor(id primitive.ObjectID) string {
	return base64.StdEncoding.EncodeToString([]byte(id.Hex()))
}

// decodeCursor decodes a base64 cursor string to an ObjectID
func decodeCursor(cursor string) (primitive.ObjectID, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid cursor format: %w", err)
	}

	objectID, err := primitive.ObjectIDFromHex(string(decoded))
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid cursor value: %w", err)
	}

	return objectID, nil
}

// PaginationParams holds pagination parameters
type PaginationParams struct {
	First  *int64
	After  *string
	Last   *int64
	Before *string
}

// PaginationResult holds paginated results and metadata
type PaginationResult struct {
	Items    []interface{}
	PageInfo PageInfo
}

// PaginateForward performs forward pagination (first + after)
// Returns up to 'first' items after the 'after' cursor
func PaginateForward(ctx context.Context, collection *mongo.Collection, filter bson.M, first int64, after *string, sortField string) (*mongo.Cursor, error) {
	// Add cursor filter if provided
	if after != nil && *after != "" {
		cursorID, err := decodeCursor(*after)
		if err != nil {
			return nil, newInvalidInputError("Invalid cursor format")
		}
		filter["_id"] = bson.M{"$gt": cursorID}
	}

	// Fetch one extra item to determine hasNextPage
	limit := first + 1

	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: sortField, Value: 1}})

	return collection.Find(ctx, filter, opts)
}

// PaginateBackward performs backward pagination (last + before)
// Returns up to 'last' items before the 'before' cursor
func PaginateBackward(ctx context.Context, collection *mongo.Collection, filter bson.M, last int64, before *string, sortField string) (*mongo.Cursor, error) {
	// Add cursor filter if provided
	if before != nil && *before != "" {
		cursorID, err := decodeCursor(*before)
		if err != nil {
			return nil, newInvalidInputError("Invalid cursor format")
		}
		filter["_id"] = bson.M{"$lt": cursorID}
	}

	// Fetch one extra item to determine hasPreviousPage
	limit := last + 1

	// Sort in reverse order for backward pagination
	opts := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: sortField, Value: -1}})

	return collection.Find(ctx, filter, opts)
}

// BuildPageInfo constructs PageInfo from pagination results
func BuildPageInfo(items []interface{}, requested int64, hasMore bool, isForward bool) PageInfo {
	pageInfo := PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
	}

	if len(items) == 0 {
		return pageInfo
	}

	if isForward {
		// Forward pagination: hasNextPage if we got more items than requested
		pageInfo.HasNextPage = hasMore

		// Start cursor is the first item's ID
		if firstItem, ok := items[0].(map[string]interface{}); ok {
			if id, ok := firstItem["_id"].(primitive.ObjectID); ok {
				cursor := encodeCursor(id)
				pageInfo.StartCursor = &cursor
			}
		}

		// End cursor is the last item's ID
		if lastItem, ok := items[len(items)-1].(map[string]interface{}); ok {
			if id, ok := lastItem["_id"].(primitive.ObjectID); ok {
				cursor := encodeCursor(id)
				pageInfo.EndCursor = &cursor
			}
		}
	} else {
		// Backward pagination: hasPreviousPage if we got more items than requested
		pageInfo.HasPreviousPage = hasMore

		// For backward pagination, items are in reverse order
		// Start cursor is the last item (which is chronologically first)
		if len(items) > 0 {
			if lastItem, ok := items[len(items)-1].(map[string]interface{}); ok {
				if id, ok := lastItem["_id"].(primitive.ObjectID); ok {
					cursor := encodeCursor(id)
					pageInfo.StartCursor = &cursor
				}
			}
		}

		// End cursor is the first item (which is chronologically last)
		if firstItem, ok := items[0].(map[string]interface{}); ok {
			if id, ok := firstItem["_id"].(primitive.ObjectID); ok {
				cursor := encodeCursor(id)
				pageInfo.EndCursor = &cursor
			}
		}
	}

	return pageInfo
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(params PaginationParams) error {
	// Cannot use both first and last
	if params.First != nil && params.Last != nil {
		return newInvalidInputError("Cannot specify both 'first' and 'last' parameters")
	}

	// Cannot use after with last
	if params.After != nil && params.Last != nil {
		return newInvalidInputError("Cannot use 'after' with 'last' parameter")
	}

	// Cannot use before with first
	if params.Before != nil && params.First != nil {
		return newInvalidInputError("Cannot use 'before' with 'first' parameter")
	}

	// Validate first is positive
	if params.First != nil && *params.First <= 0 {
		return newInvalidInputError("'first' parameter must be positive")
	}

	// Validate last is positive
	if params.Last != nil && *params.Last <= 0 {
		return newInvalidInputError("'last' parameter must be positive")
	}

	// Validate maximum page size (1000 items as per spec)
	maxPageSize := int64(1000)
	if params.First != nil && *params.First > maxPageSize {
		return newInvalidInputError(fmt.Sprintf("'first' parameter cannot exceed %d", maxPageSize))
	}
	if params.Last != nil && *params.Last > maxPageSize {
		return newInvalidInputError(fmt.Sprintf("'last' parameter cannot exceed %d", maxPageSize))
	}

	return nil
}
