package db

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection interface defines operations on a MongoDB collection (T057)
// Provides a generic persistence interface for CRUD operations
type Collection interface {
	// InsertOne inserts a single document
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)

	// InsertMany inserts multiple documents
	InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error)

	// FindOne finds a single document matching the filter
	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult

	// Find finds multiple documents matching the filter
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)

	// UpdateOne updates a single document matching the filter
	UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)

	// UpdateMany updates multiple documents matching the filter
	UpdateMany(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)

	// DeleteOne deletes a single document matching the filter
	DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)

	// DeleteMany deletes multiple documents matching the filter
	DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)

	// CountDocuments counts documents matching the filter
	CountDocuments(ctx context.Context, filter interface{}) (int64, error)

	// Name returns the collection name
	Name() string
}

// collectionWrapper wraps mongo.Collection with timeout and logging (T058)
type collectionWrapper struct {
	collection       *mongo.Collection
	name             string
	operationTimeout time.Duration // Default timeout for operations (5-10s per FR-007)
	logger           zerolog.Logger
}

// newCollection creates a new collection wrapper (T059)
func newCollection(coll *mongo.Collection, operationTimeout time.Duration, logger zerolog.Logger) Collection {
	return &collectionWrapper{
		collection:       coll,
		name:             coll.Name(),
		operationTimeout: operationTimeout,
		logger:           logger,
	}
}

// Name returns the collection name (T069)
func (c *collectionWrapper) Name() string {
	return c.name
}

// withTimeout creates a context with operation timeout if not already set (T070)
func (c *collectionWrapper) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	// If context already has a deadline, use it
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}

	// Apply default operation timeout (5-10s per FR-007, FR-018)
	return context.WithTimeout(ctx, c.operationTimeout)
}

// InsertOne inserts a single document (T060)
func (c *collectionWrapper) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result, err := c.collection.InsertOne(ctx, document)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "insert_one").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Insert operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "insert_one").
		Str("collection", c.name).
		Dur("duration_ms", duration).
		Interface("inserted_id", result.InsertedID).
		Msg("Document inserted")

	return result, nil
}

// InsertMany inserts multiple documents (T061)
func (c *collectionWrapper) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result, err := c.collection.InsertMany(ctx, documents)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "insert_many").
			Str("collection", c.name).
			Int("document_count", len(documents)).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Batch insert operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "insert_many").
		Str("collection", c.name).
		Int("inserted_count", len(result.InsertedIDs)).
		Dur("duration_ms", duration).
		Msg("Documents inserted")

	return result, nil
}

// FindOne finds a single document (T062)
func (c *collectionWrapper) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result := c.collection.FindOne(ctx, filter)

	duration := time.Since(startTime)

	// Check for errors (ErrNotFound is common and not logged as error)
	err := result.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		c.logger.Error().
			Str("operation", "find_one").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Find operation failed")
	} else if err == mongo.ErrNoDocuments {
		c.logger.Debug().
			Str("operation", "find_one").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Msg("Document not found")
	} else {
		c.logger.Debug().
			Str("operation", "find_one").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Msg("Document found")
	}

	return result
}

// Find finds multiple documents (T063)
func (c *collectionWrapper) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	cursor, err := c.collection.Find(ctx, filter, opts...)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "find").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Find operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "find").
		Str("collection", c.name).
		Dur("duration_ms", duration).
		Msg("Find operation completed")

	return cursor, nil
}

// UpdateOne updates a single document (T064)
func (c *collectionWrapper) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result, err := c.collection.UpdateOne(ctx, filter, update)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "update_one").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Update operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "update_one").
		Str("collection", c.name).
		Int64("matched_count", result.MatchedCount).
		Int64("modified_count", result.ModifiedCount).
		Dur("duration_ms", duration).
		Msg("Document updated")

	return result, nil
}

// UpdateMany updates multiple documents (T065)
func (c *collectionWrapper) UpdateMany(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result, err := c.collection.UpdateMany(ctx, filter, update)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "update_many").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Batch update operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "update_many").
		Str("collection", c.name).
		Int64("matched_count", result.MatchedCount).
		Int64("modified_count", result.ModifiedCount).
		Dur("duration_ms", duration).
		Msg("Documents updated")

	return result, nil
}

// DeleteOne deletes a single document (T066)
func (c *collectionWrapper) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result, err := c.collection.DeleteOne(ctx, filter)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "delete_one").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Delete operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "delete_one").
		Str("collection", c.name).
		Int64("deleted_count", result.DeletedCount).
		Dur("duration_ms", duration).
		Msg("Document deleted")

	return result, nil
}

// DeleteMany deletes multiple documents (T067)
func (c *collectionWrapper) DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	result, err := c.collection.DeleteMany(ctx, filter)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "delete_many").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Batch delete operation failed")
		return nil, err
	}

	c.logger.Debug().
		Str("operation", "delete_many").
		Str("collection", c.name).
		Int64("deleted_count", result.DeletedCount).
		Dur("duration_ms", duration).
		Msg("Documents deleted")

	return result, nil
}

// CountDocuments counts documents matching the filter (T068)
func (c *collectionWrapper) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	count, err := c.collection.CountDocuments(ctx, filter)

	duration := time.Since(startTime)

	// Structured logging (FR-017)
	if err != nil {
		c.logger.Error().
			Str("operation", "count_documents").
			Str("collection", c.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Count operation failed")
		return 0, err
	}

	c.logger.Debug().
		Str("operation", "count_documents").
		Str("collection", c.name).
		Int64("count", count).
		Dur("duration_ms", duration).
		Msg("Documents counted")

	return count, nil
}
