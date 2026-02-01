package resolvers

import (
	"context"
	"fmt"

	"github.com/yourusername/air-go/internal/graphql/generated"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// EntityConfig defines configuration for generic entity queries
// T005: EntityConfig struct for parameterized entity queries
// T007: Added FilterConverter for search functionality
type EntityConfig struct {
	CollectionName  string                              // MongoDB collection name
	DeletionField   string                              // Field indicating deletion status (e.g., "status.deletion" or "actionIndicator")
	DeletionValue   string                              // Value indicating deleted entity (e.g., "DELETED" or "DELETE")
	SorterConverter func(interface{}) []bson.M          // Converts GraphQL sorter input to MongoDB aggregation pipeline stages
	FilterConverter func(interface{}) bson.M            // Converts GraphQL filter input to MongoDB filter (T007)
}

// T013: Entity configuration map with all 6 entities
var entityConfigs = map[string]EntityConfig{
	"customer": {
		CollectionName:  "customers",
		DeletionField:   "status.deletion",
		DeletionValue:   "DELETED",
		SorterConverter: customerSorterConverter,
		FilterConverter: func(filter interface{}) bson.M {
			if f, ok := filter.(*generated.CustomerQueryFilterInput); ok {
				return convertCustomerFilter(f)
			}
			return bson.M{}
		},
	},
	"employee": {
		CollectionName:  "employees",
		DeletionField:   "status.deletion",
		DeletionValue:   "DELETED",
		SorterConverter: employeeSorterConverter,
		FilterConverter: func(filter interface{}) bson.M {
			if f, ok := filter.(*generated.EmployeeQueryFilterInput); ok {
				return convertEmployeeFilter(f)
			}
			return bson.M{}
		},
	},
	"team": {
		CollectionName:  "teams",
		DeletionField:   "status.deletion",
		DeletionValue:   "DELETED",
		SorterConverter: teamSorterConverter, // T044: Added team sorter converter
		FilterConverter: func(filter interface{}) bson.M {
			if f, ok := filter.(*generated.TeamQueryFilterInput); ok {
				return convertTeamFilter(f)
			}
			return bson.M{}
		},
	},
	"inventory": {
		CollectionName:  "inventories",
		DeletionField:   "actionIndicator",
		DeletionValue:   "DELETE",
		SorterConverter: inventorySorterConverter,
		FilterConverter: nil, // No search functionality for inventory in this feature
	},
	"executionPlan": {
		CollectionName:  "executionPlans",
		DeletionField:   "actionIndicator",
		DeletionValue:   "DELETE",
		SorterConverter: executionPlanSorterConverter, // T044: Added execution plan sorter converter
		FilterConverter: func(filter interface{}) bson.M {
			if f, ok := filter.(*generated.ExecutionPlanQueryFilterInput); ok {
				return convertExecutionPlanFilter(f)
			}
			return bson.M{}
		},
	},
	"referencePortfolio": {
		CollectionName:  "referencePortfolios",
		DeletionField:   "actionIndicator",
		DeletionValue:   "DELETE",
		SorterConverter: referencePortfolioSorterConverter, // T044: Added reference portfolio sorter converter
		FilterConverter: func(filter interface{}) bson.M {
			if f, ok := filter.(*generated.ReferencePortfolioQueryFilterInput); ok {
				return convertReferencePortfolioFilter(f)
			}
			return bson.M{}
		},
	},
}

// T006: UUID validation helper function (using existing isValidUUID from customer.go)

// T007: Batch size validation helper function
func validateBatchSizeGeneric(identifiers []string) error {
	if len(identifiers) > MaxBatchSize {
		return newInvalidInputError(fmt.Sprintf(
			"batch size exceeds maximum: requested %d, maximum %d",
			len(identifiers),
			MaxBatchSize,
		))
	}
	return nil
}

// T008: Identifier deduplication helper function
func deduplicateIdentifiersGeneric(identifiers []string) []string {
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

// T011: Convert SortEnumType to MongoDB sort direction integer
func sortEnumToInt(sortEnum generated.SortEnumType) int {
	if sortEnum == generated.SortEnumTypeAsc {
		return 1
	}
	return -1
}

// T012: Append null-safe sorting stages for SQL-standard null handling
// ASC: non-nulls first (ascending), nulls last
// DESC: nulls first, non-nulls last (descending)
func appendNullSafeSorting(pipeline []bson.M, field string, sortEnum generated.SortEnumType) []bson.M {
	if sortEnum == generated.SortEnumTypeAsc {
		// For ascending: non-nulls first, nulls last
		pipeline = append(pipeline, bson.M{
			"$addFields": bson.M{
				"_sortKey": bson.M{
					"$ifNull": []interface{}{
						"$" + field,
						"zzzzzzz-null-placeholder", // Sorts after all valid values
					},
				},
			},
		})
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"_sortKey": 1}})
		pipeline = append(pipeline, bson.M{"$project": bson.M{"_sortKey": 0}}) // Remove temp field
	} else {
		// For descending: nulls first, non-nulls last
		pipeline = append(pipeline, bson.M{
			"$addFields": bson.M{
				"_sortKey": bson.M{
					"$ifNull": []interface{}{
						"$" + field,
						"zzzzzzz-null-placeholder", // Sorts first when descending
					},
				},
			},
		})
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"_sortKey": -1}})
		pipeline = append(pipeline, bson.M{"$project": bson.M{"_sortKey": 0}})
	}

	return pipeline
}

// T014: Structured logging helper exists in logging.go - using that implementation

// T009: Generic getEntity function for single entity retrieval
// Retrieves a single entity by identifier, excluding deleted entities
// Returns nil if entity not found or deleted
func getEntity(ctx context.Context, dbClient interface{}, config EntityConfig, identifier string, result interface{}) error {
	// Validate UUID format
	if !isValidUUID(identifier) {
		return newInvalidInputError("invalid UUID format")
	}

	// Cast to DBClient interface
	db, ok := dbClient.(DBClient)
	if !ok {
		return &QueryError{
			Message: "Database not available",
			Code:    ErrCodeDatabaseError,
		}
	}

	// Get collection
	collection := db.Collection(config.CollectionName)

	// Build query filter: match identifier and exclude deleted entities
	filter := bson.M{
		"identifier":         identifier,
		config.DeletionField: bson.M{"$ne": config.DeletionValue},
	}

	// Execute FindOne query
	findResult := collection.FindOne(ctx, filter)
	if findResult.Err() == mongo.ErrNoDocuments {
		// Entity not found or deleted - return nil (result will have zero values)
		return nil
	}
	if findResult.Err() != nil {
		return mapMongoError(findResult.Err())
	}

	if decodeErr := findResult.Decode(result); decodeErr != nil {
		return mapMongoError(decodeErr)
	}

	return nil
}

// T010: Generic getEntitiesByKeys function for batch entity retrieval
// Retrieves multiple entities by identifiers with optional ordering
// Returns empty array if no identifiers provided or no matches found
func getEntitiesByKeys(ctx context.Context, dbClient interface{}, config EntityConfig, identifiers []string, sorter interface{}, result interface{}) error {
	// Validate batch size
	if err := validateBatchSizeGeneric(identifiers); err != nil {
		return err
	}

	// Handle empty array case
	if len(identifiers) == 0 {
		// result should already be initialized as empty slice by caller
		return nil
	}

	// Validate all UUID formats
	for _, id := range identifiers {
		if !isValidUUID(id) {
			return newInvalidInputError(fmt.Sprintf("invalid UUID format: %s", id))
		}
	}

	// Deduplicate identifiers
	dedupedIDs := deduplicateIdentifiersGeneric(identifiers)

	// Build base aggregation pipeline
	pipeline := []bson.M{
		{"$match": bson.M{
			"identifier":         bson.M{"$in": dedupedIDs},
			config.DeletionField: bson.M{"$ne": config.DeletionValue},
		}},
	}

	// Apply entity-specific sorting if sorter converter exists and sorter is provided
	if config.SorterConverter != nil && sorter != nil {
		sortStages := config.SorterConverter(sorter)
		pipeline = append(pipeline, sortStages...)
	} else {
		// Default sorting by identifier ascending
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	// Cast to DBClient interface
	db, ok := dbClient.(DBClient)
	if !ok {
		return &QueryError{
			Message: "Database not available",
			Code:    ErrCodeDatabaseError,
		}
	}

	// Get collection
	collection := db.Collection(config.CollectionName)

	// Execute aggregation pipeline
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return &QueryError{
			Message: "Database query failed",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}
	defer cursor.Close(ctx)

	// Decode all results
	if err := cursor.All(ctx, result); err != nil {
		return &QueryError{
			Message: "Failed to decode entities",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}

	return nil
}

// T057: Customer sorter converter
func customerSorterConverter(sorter interface{}) []bson.M {
	s, ok := sorter.([]*generated.CustomerQuerySorterInput)
	if !ok || len(s) == 0 {
		return []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	sortSpec := s[0]
	pipeline := []bson.M{}

	// Map each GraphQL sorter field to MongoDB sort stage
	if sortSpec.FirstName != nil {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"firstName": sortEnumToInt(*sortSpec.FirstName)}})
	}

	if sortSpec.LastName != nil {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"lastName": sortEnumToInt(*sortSpec.LastName)}})
	}

	if sortSpec.BirthDate != nil {
		pipeline = appendNullSafeSorting(pipeline, "birthDate", *sortSpec.BirthDate)
	}

	if sortSpec.EmployeeEmail != nil {
		pipeline = appendNullSafeSorting(pipeline, "employeeEmail", *sortSpec.EmployeeEmail)
	}

	if sortSpec.Payment != nil && sortSpec.Payment.Status != nil {
		pipeline = appendNullSafeSorting(pipeline, "payment.status", *sortSpec.Payment.Status)
	}

	if sortSpec.CreateDate != nil {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"createDate": sortEnumToInt(*sortSpec.CreateDate)}})
	}

	// Default to identifier if no fields specified
	if len(pipeline) == 0 {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	return pipeline
}

// T058: Employee sorter converter
func employeeSorterConverter(sorter interface{}) []bson.M {
	s, ok := sorter.([]*generated.EmployeeQuerySorterInput)
	if !ok || len(s) == 0 {
		return []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	sortSpec := s[0]
	pipeline := []bson.M{}

	if sortSpec.FirstName != nil {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"firstName": sortEnumToInt(*sortSpec.FirstName)}})
	}

	if sortSpec.LastName != nil {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"lastName": sortEnumToInt(*sortSpec.LastName)}})
	}

	if sortSpec.BirthDate != nil {
		pipeline = appendNullSafeSorting(pipeline, "birthDate", *sortSpec.BirthDate)
	}

	if sortSpec.UserEmail != nil {
		pipeline = appendNullSafeSorting(pipeline, "userEmail", *sortSpec.UserEmail)
	}

	// Default to identifier if no fields specified
	if len(pipeline) == 0 {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	return pipeline
}

// T059: Inventory sorter converter
func inventorySorterConverter(sorter interface{}) []bson.M {
	s, ok := sorter.([]*generated.InventoryQuerySorterInput)
	if !ok || len(s) == 0 {
		return []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	sortSpec := s[0]
	pipeline := []bson.M{}

	if sortSpec.CustomerID != nil {
		pipeline = appendNullSafeSorting(pipeline, "customerId", *sortSpec.CustomerID)
	}

	// Default to identifier if no fields specified
	if len(pipeline) == 0 {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	return pipeline
}

// T041: Team sorter converter
func teamSorterConverter(sorter interface{}) []bson.M {
	s, ok := sorter.([]*generated.TeamQuerySorterInput)
	if !ok || len(s) == 0 {
		return []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	// Build a single $sort document with all fields
	sortDoc := bson.M{}

	// Process all sorter inputs in order
	for _, sortSpec := range s {
		if sortSpec.Name != nil {
			sortDoc["name"] = sortEnumToInt(*sortSpec.Name)
		}

		if sortSpec.Description != nil {
			sortDoc["description"] = sortEnumToInt(*sortSpec.Description)
		}

		if sortSpec.IsShared != nil {
			sortDoc["isShared"] = sortEnumToInt(*sortSpec.IsShared)
		}

		if sortSpec.EmployeeID != nil {
			sortDoc["employeeId"] = sortEnumToInt(*sortSpec.EmployeeID)
		}
	}

	// Default to identifier if no fields specified
	if len(sortDoc) == 0 {
		sortDoc["identifier"] = 1
	}

	return []bson.M{{"$sort": sortDoc}}
}

// T042: ExecutionPlan sorter converter
func executionPlanSorterConverter(sorter interface{}) []bson.M {
	s, ok := sorter.([]*generated.ExecutionPlanQuerySorterInput)
	if !ok || len(s) == 0 {
		return []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	pipeline := []bson.M{}

	// Process all sorter inputs in order
	for _, sortSpec := range s {
		if sortSpec.CustomerID != nil {
			pipeline = appendNullSafeSorting(pipeline, "customerId", *sortSpec.CustomerID)
		}
	}

	// Default to identifier if no fields specified
	if len(pipeline) == 0 {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	return pipeline
}

// T043: ReferencePortfolio sorter converter
func referencePortfolioSorterConverter(sorter interface{}) []bson.M {
	s, ok := sorter.([]*generated.ReferencePortfolioQuerySorterInput)
	if !ok || len(s) == 0 {
		return []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	pipeline := []bson.M{}

	// Process all sorter inputs in order
	for _, sortSpec := range s {
		if sortSpec.CustomerID != nil {
			pipeline = appendNullSafeSorting(pipeline, "customerId", *sortSpec.CustomerID)
		}
	}

	// Default to identifier if no fields specified
	if len(pipeline) == 0 {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"identifier": 1}})
	}

	return pipeline
}
