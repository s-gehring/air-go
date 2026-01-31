package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"go.mongodb.org/mongo-driver/bson"
)

// T018: Batch size validation
func validateBatchSize(identifiers []string) error {
	if len(identifiers) > MaxBatchSize {
		return newInvalidInputError(fmt.Sprintf(
			"batch size exceeds maximum: requested %d, maximum %d",
			len(identifiers),
			MaxBatchSize,
		))
	}
	return nil
}

// T019: UUID format validation for all identifiers
func validateUUIDs(identifiers []string) error {
	for _, id := range identifiers {
		if !isValidUUID(id) {
			return newInvalidInputError(fmt.Sprintf("invalid UUID format: %s", id))
		}
	}
	return nil
}

// T020: De-duplication using map
func deduplicateIdentifiers(identifiers []string) []string {
	seen := make(map[string]bool)
	deduped := make([]string, 0, len(identifiers))

	for _, id := range identifiers {
		if !seen[id] {
			seen[id] = true
			deduped = append(deduped, id)
		}
	}

	return deduped
}

// T021: Build MongoDB filter with $in operator and deletion status check
func buildInventoryFilter(identifiers []string) bson.M {
	return bson.M{
		"identifier": bson.M{"$in": identifiers},
		"actionIndicator": bson.M{"$ne": "DELETE"},
	}
}

// T022: Build aggregation pipeline with ordering
func buildInventoryPipeline(identifiers []string, order []*generated.InventoryQuerySorterInput) []bson.M {
	filter := buildInventoryFilter(identifiers)

	pipeline := []bson.M{
		{"$match": filter},
	}

	// Apply ordering if specified
	if order != nil && len(order) > 0 {
		sortSpec := order[0]
		if sortSpec.CustomerID != nil {
			pipeline = appendCustomerIDSorting(pipeline, *sortSpec.CustomerID)
		}
	} else {
		// Default ordering: identifier ascending
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	return pipeline
}

// Helper function for SQL-standard null handling in customerID ordering
func appendCustomerIDSorting(pipeline []bson.M, sortDirection generated.SortEnumType) []bson.M {
	// SQL-standard null handling:
	// ASC: nulls last
	// DESC: nulls first

	if sortDirection == generated.SortEnumTypeAsc {
		// For ascending: non-nulls first (ascending), nulls last
		pipeline = append(pipeline, bson.M{
			"$addFields": bson.M{
				"_sortKey": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": []interface{}{"$customerId", nil}},
						"then": "zzzzzzz-null-placeholder", // Sorts after all valid UUIDs
						"else": "$customerId",
					},
				},
			},
		})
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"_sortKey": 1}})
		pipeline = append(pipeline, bson.M{"$project": bson.M{"_sortKey": 0}}) // Remove temp field
	} else {
		// For descending: nulls first, non-nulls last (descending)
		pipeline = append(pipeline, bson.M{
			"$addFields": bson.M{
				"_sortKey": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": []interface{}{"$customerId", nil}},
						"then": "0000000-null-placeholder", // Sorts before all valid UUIDs
						"else": "$customerId",
					},
				},
			},
		})
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"_sortKey": -1}})
		pipeline = append(pipeline, bson.M{"$project": bson.M{"_sortKey": 0}})
	}

	return pipeline
}

// T023: Fetch inventories from database
func (r *queryResolver) fetchInventories(ctx context.Context, pipeline []bson.M) ([]*generated.Inventory, error) {
	collection := r.DBClient.Collection("inventories")
	if collection == nil {
		return nil, &QueryError{
			Message: "Database not available",
			Code:    ErrCodeDatabaseError,
		}
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, &QueryError{
			Message: "Database query failed",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}
	defer cursor.Close(ctx)

	var inventories []*generated.Inventory
	if err := cursor.All(ctx, &inventories); err != nil {
		return nil, &QueryError{
			Message: "Failed to decode inventories",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}

	return inventories, nil
}

// T024: ByKeysGet resolver function integrating all logic
// T025: Error handling with standard GraphQL error format
// T026-T029: Structured logging per OBS-001 through OBS-004
func (r *queryResolver) ByKeysGet(
	ctx context.Context,
	identifiers []string,
	order []*generated.InventoryQuerySorterInput,
) ([]*generated.Inventory, error) {
	// T026: Log query parameters per OBS-001
	startTime := time.Now()
	identifierCount := len(identifiers)

	var orderStr string
	if order != nil && len(order) > 0 && order[0].CustomerID != nil {
		orderStr = fmt.Sprintf("customerId %s", *order[0].CustomerID)
	} else {
		orderStr = "identifier ASC (default)"
	}

	log.Info().
		Int("identifierCount", identifierCount).
		Str("order", orderStr).
		Str("query", "byKeysGet").
		Msg("byKeysGet query started")

	var resultCount int
	var err error

	// T027: Log query duration per OBS-002
	// T029: Log errors per OBS-004
	defer func() {
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			// Log error case
			log.Error().
				Err(err).
				Int("identifierCount", identifierCount).
				Int64("duration", duration).
				Str("query", "byKeysGet").
				Msg("byKeysGet query failed")
		} else {
			// T028: Log result count per OBS-003
			log.Info().
				Int("identifierCount", identifierCount).
				Int("resultCount", resultCount).
				Int64("duration", duration).
				Str("query", "byKeysGet").
				Msg("byKeysGet query completed")
		}
	}()

	// T018: Validate batch size
	if err = validateBatchSize(identifiers); err != nil {
		return nil, err
	}

	// Handle empty array case
	if len(identifiers) == 0 {
		resultCount = 0
		return []*generated.Inventory{}, nil
	}

	// T019: Validate UUID formats
	if err = validateUUIDs(identifiers); err != nil {
		return nil, err
	}

	// T020: De-duplicate identifiers
	dedupedIDs := deduplicateIdentifiers(identifiers)

	// T022: Build aggregation pipeline
	pipeline := buildInventoryPipeline(dedupedIDs, order)

	// T023: Fetch inventories from database
	inventories, err := r.fetchInventories(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	resultCount = len(inventories)
	return inventories, nil
}
