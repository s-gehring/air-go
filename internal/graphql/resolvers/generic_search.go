package resolvers

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// T006: Generic searchEntities function for entity search with filtering, sorting, and pagination
// T009: Validation helpers for pagination parameters

// validatePaginationParams validates first/last pagination parameters
// Returns error if both first and last are specified, or if limits exceed MaxBatchSize
func validatePaginationParams(first, last *int) error {
	// Cannot specify both forward and backward pagination
	if first != nil && last != nil {
		return newInvalidInputError("cannot specify both 'first' and 'last' pagination parameters")
	}

	// Validate first parameter
	if first != nil {
		if *first < 0 {
			return newInvalidInputError("'first' must be non-negative")
		}
		if *first > MaxBatchSize {
			return newInvalidInputError(fmt.Sprintf("'first' exceeds maximum batch size: requested %d, maximum %d", *first, MaxBatchSize))
		}
	}

	// Validate last parameter
	if last != nil {
		if *last < 0 {
			return newInvalidInputError("'last' must be non-negative")
		}
		if *last > MaxBatchSize {
			return newInvalidInputError(fmt.Sprintf("'last' exceeds maximum batch size: requested %d, maximum %d", *last, MaxBatchSize))
		}
	}

	return nil
}

// buildPaginationFilter builds a MongoDB filter for cursor-based pagination
// The filter ensures we only get documents after/before the cursor position
// Based on sort fields and identifier in the cursor
func buildPaginationFilter(cursor *Cursor, sortFields []string, isForward bool) bson.M {
	if cursor == nil || len(cursor.SortFields) == 0 {
		return bson.M{}
	}

	// Build $or conditions for pagination
	// For cursor at position [value1, value2, identifier]:
	// Forward (after): field1 > value1 OR (field1 = value1 AND field2 > value2) OR (field1 = value1 AND field2 = value2 AND identifier > cursorId)
	// Backward (before): Similar but with < operators

	orConditions := []bson.M{}

	// Determine comparison operator based on direction
	gtOp := "$gt"
	if !isForward {
		gtOp = "$lt"
	}

	// Build cascading OR conditions
	for i := 0; i < len(sortFields); i++ {
		condition := bson.M{}

		// All previous fields must equal cursor values
		for j := 0; j < i; j++ {
			if j < len(cursor.SortFields) {
				condition[sortFields[j]] = cursor.SortFields[j]
			}
		}

		// Current field must be greater/less than cursor value
		if i < len(cursor.SortFields) {
			condition[sortFields[i]] = bson.M{gtOp: cursor.SortFields[i]}
		}

		orConditions = append(orConditions, condition)
	}

	// Final condition: all sort fields equal, identifier greater/less than cursor identifier
	finalCondition := bson.M{}
	for i, field := range sortFields {
		if i < len(cursor.SortFields) {
			finalCondition[field] = cursor.SortFields[i]
		}
	}
	finalCondition["identifier"] = bson.M{gtOp: cursor.Identifier}
	orConditions = append(orConditions, finalCondition)

	if len(orConditions) == 0 {
		return bson.M{}
	}

	return bson.M{"$or": orConditions}
}

// searchEntities performs generic entity search with filtering, sorting, and pagination
// Returns count, data array, totalCount, and pagination info
func searchEntities(
	ctx context.Context,
	dbClient interface{},
	config EntityConfig,
	filter interface{}, // Entity-specific filter (converted to bson.M by FilterConverter)
	sorter interface{}, // Entity-specific sorter (converted to pipeline stages by SorterConverter)
	first *int, after *string, last *int, before *string, // Pagination parameters
	result interface{}, // Pointer to slice of entity type (will be populated with decoded results)
) (count int, totalCount int, hasNextPage bool, hasPreviousPage bool, startCursor *string, endCursor *string, err error) {
	// Validate pagination parameters
	if err := validatePaginationParams(first, last); err != nil {
		return 0, 0, false, false, nil, nil, err
	}

	// Determine effective limit
	effectiveLimit := MaxBatchSize
	if first != nil && *first > 0 {
		effectiveLimit = *first
	} else if last != nil && *last > 0 {
		effectiveLimit = *last
	}

	// Decode cursors if provided
	var afterCursor *Cursor
	var beforeCursor *Cursor

	if after != nil && *after != "" {
		afterCursor, err = decodeCursor(*after)
		if err != nil {
			return 0, 0, false, false, nil, nil, err
		}
	}

	if before != nil && *before != "" {
		beforeCursor, err = decodeCursor(*before)
		if err != nil {
			return 0, 0, false, false, nil, nil, err
		}
	}

	// Build base filter (deletion exclusion + entity filter)
	baseFilter := bson.M{
		config.DeletionField: bson.M{"$ne": config.DeletionValue},
	}

	// Apply entity-specific filter if FilterConverter exists and filter is provided
	if config.FilterConverter != nil && filter != nil {
		entityFilter := config.FilterConverter(filter)
		if len(entityFilter) > 0 {
			// Combine deletion filter with entity filter using $and
			baseFilter = bson.M{
				"$and": []bson.M{
					{config.DeletionField: bson.M{"$ne": config.DeletionValue}},
					entityFilter,
				},
			}
		}
	}

	// Build aggregation pipeline
	pipeline := []bson.M{
		{"$match": baseFilter},
	}

	// Apply sorting
	var sortStages []bson.M
	if config.SorterConverter != nil && sorter != nil {
		sortStages = config.SorterConverter(sorter)
	} else {
		// Default sorting by identifier ascending
		sortStages = []bson.M{{"$sort": bson.M{"identifier": 1}}}
	}

	// For pagination filter, we need to know the sort field names
	// Extract from sort stages
	var sortFieldNames []string
	if len(sortStages) > 0 {
		for _, stage := range sortStages {
			if sortSpec, ok := stage["$sort"].(bson.M); ok {
				for fieldName := range sortSpec {
					if fieldName != "_sortKey" { // Skip temporary sort keys
						sortFieldNames = append(sortFieldNames, fieldName)
					}
				}
			}
		}
	}

	// Use $facet to get both count and paginated data in a single query
	facetPipeline := bson.M{
		"$facet": bson.M{
			"metadata": []bson.M{
				{"$count": "totalCount"},
			},
			"data": buildDataPipeline(sortStages, afterCursor, beforeCursor, sortFieldNames, first, last, effectiveLimit),
		},
	}

	pipeline = append(pipeline, facetPipeline)

	// Execute aggregation
	db, ok := dbClient.(DBClient)
	if !ok {
		return 0, 0, false, false, nil, nil, &QueryError{
			Message: "Database not available",
			Code:    ErrCodeDatabaseError,
		}
	}

	collection := db.Collection(config.CollectionName)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, false, false, nil, nil, &QueryError{
			Message: "Database query failed",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}
	defer cursor.Close(ctx)

	// Parse facet results
	var facetResults []struct {
		Metadata []struct {
			TotalCount int `bson:"totalCount"`
		} `bson:"metadata"`
		Data []bson.Raw `bson:"data"` // Use bson.Raw for flexible decoding
	}

	if err := cursor.All(ctx, &facetResults); err != nil {
		return 0, 0, false, false, nil, nil, &QueryError{
			Message: "Failed to decode search results",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}

	// Handle empty results
	if len(facetResults) == 0 {
		return 0, 0, false, false, nil, nil, nil
	}

	facetResult := facetResults[0]

	// Get totalCount
	if len(facetResult.Metadata) > 0 {
		totalCount = facetResult.Metadata[0].TotalCount
	}

	// Decode data into result slice
	dataCount := len(facetResult.Data)

	// Handle empty data
	if dataCount == 0 {
		return 0, totalCount, false, false, nil, nil, nil
	}

	// Determine if we have extra items for pagination detection
	isForward := first != nil || (first == nil && last == nil)

	if isForward {
		// Forward pagination: check if we got limit+1 items
		if dataCount > effectiveLimit {
			hasNextPage = true
			// Trim to effectiveLimit
			facetResult.Data = facetResult.Data[:effectiveLimit]
			dataCount = effectiveLimit
		}
		hasPreviousPage = afterCursor != nil
	} else {
		// Backward pagination: check if we got limit+1 items
		if dataCount > effectiveLimit {
			hasPreviousPage = true
			// Trim first item (we queried in reverse)
			facetResult.Data = facetResult.Data[1:]
			dataCount = effectiveLimit
		}
		hasNextPage = beforeCursor != nil
	}

	// Decode trimmed data into result
	dataBytes := make([]byte, 0)
	for _, raw := range facetResult.Data {
		dataBytes = append(dataBytes, raw...)
	}

	// Use bson.Unmarshal to decode into the result slice
	// Create a temporary array to decode all items
	tempArray := make([]bson.M, len(facetResult.Data))
	for i, raw := range facetResult.Data {
		if err := bson.Unmarshal(raw, &tempArray[i]); err != nil {
			return 0, 0, false, false, nil, nil, &QueryError{
				Message: "Failed to decode entity data",
				Code:    ErrCodeDatabaseError,
				Cause:   err,
			}
		}
	}

	// Marshal temp array and unmarshal into result type
	tempBytes, err := bson.Marshal(bson.M{"items": tempArray})
	if err != nil {
		return 0, 0, false, false, nil, nil, &QueryError{
			Message: "Failed to encode entity data",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}

	var wrapper struct {
		Items interface{} `bson:"items"`
	}
	wrapper.Items = result

	if err := bson.Unmarshal(tempBytes, &wrapper); err != nil {
		return 0, 0, false, false, nil, nil, &QueryError{
			Message: "Failed to decode entities into result type",
			Code:    ErrCodeDatabaseError,
			Cause:   err,
		}
	}

	count = dataCount

	// Generate cursors from first and last items
	if count > 0 {
		// Start cursor: from first item
		firstItem := tempArray[0]
		startCursorValue, err := generateCursor(firstItem, sortFieldNames)
		if err == nil {
			startCursor = &startCursorValue
		}

		// End cursor: from last item
		lastItem := tempArray[count-1]
		endCursorValue, err := generateCursor(lastItem, sortFieldNames)
		if err == nil {
			endCursor = &endCursorValue
		}
	}

	return count, totalCount, hasNextPage, hasPreviousPage, startCursor, endCursor, nil
}

// generateCursor creates a cursor string from an entity document and sort fields
func generateCursor(doc bson.M, sortFieldNames []string) (string, error) {
	cursor := Cursor{
		SortFields: make([]interface{}, 0, len(sortFieldNames)),
	}

	// Extract sort field values
	for _, fieldName := range sortFieldNames {
		if fieldName == "identifier" {
			continue // Skip identifier in sort fields, we'll add it separately
		}
		value := doc[fieldName]
		cursor.SortFields = append(cursor.SortFields, value)
	}

	// Always add identifier as tiebreaker
	identifier, ok := doc["identifier"].(string)
	if !ok {
		return "", fmt.Errorf("document missing identifier field")
	}
	cursor.Identifier = identifier

	// Encode cursor
	return encodeCursor(cursor)
}

// buildDataPipeline constructs the data branch of the $facet pipeline
func buildDataPipeline(sortStages []bson.M, afterCursor, beforeCursor *Cursor, sortFieldNames []string, first, last *int, effectiveLimit int) []bson.M {
	dataPipeline := []bson.M{}

	// Apply sorting stages
	dataPipeline = append(dataPipeline, sortStages...)

	// Apply cursor-based pagination filter
	isForward := first != nil || (first == nil && last == nil)

	if isForward && afterCursor != nil {
		paginationFilter := buildPaginationFilter(afterCursor, sortFieldNames, true)
		if len(paginationFilter) > 0 {
			dataPipeline = append(dataPipeline, bson.M{"$match": paginationFilter})
		}
	} else if !isForward && beforeCursor != nil {
		paginationFilter := buildPaginationFilter(beforeCursor, sortFieldNames, false)
		if len(paginationFilter) > 0 {
			dataPipeline = append(dataPipeline, bson.M{"$match": paginationFilter})
		}
	}

	// Apply limit (+1 to detect hasNextPage/hasPreviousPage)
	dataPipeline = append(dataPipeline, bson.M{"$limit": effectiveLimit + 1})

	return dataPipeline
}
