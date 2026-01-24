package db

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database interface defines admin operations on a MongoDB database (T072)
// Used for database-level operations like listing collections and database management
type Database interface {
	// Drop drops the entire database
	Drop(ctx context.Context) error

	// CreateCollection creates a new collection with options
	CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error

	// ListCollectionNames returns names of all collections in the database
	ListCollectionNames(ctx context.Context, filter interface{}) ([]string, error)

	// Collection returns a Collection interface for the named collection
	Collection(name string) Collection

	// Name returns the database name
	Name() string
}

// databaseWrapper wraps mongo.Database with timeout and logging
type databaseWrapper struct {
	database         *mongo.Database
	name             string
	operationTimeout time.Duration
	logger           zerolog.Logger
}

// newDatabase creates a new database wrapper
func newDatabase(db *mongo.Database, operationTimeout time.Duration, logger zerolog.Logger) Database {
	return &databaseWrapper{
		database:         db,
		name:             db.Name(),
		operationTimeout: operationTimeout,
		logger:           logger,
	}
}

// Name returns the database name
func (d *databaseWrapper) Name() string {
	return d.name
}

// withTimeout creates a context with operation timeout if not already set
func (d *databaseWrapper) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	// If context already has a deadline, use it
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}

	// Apply default operation timeout
	return context.WithTimeout(ctx, d.operationTimeout)
}

// Drop drops the entire database (T074)
// WARNING: This is a destructive operation. Protected by build tags in tests.
func (d *databaseWrapper) Drop(ctx context.Context) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	err := d.database.Drop(ctx)

	duration := time.Since(startTime)

	// Structured logging
	if err != nil {
		d.logger.Error().
			Str("operation", "drop_database").
			Str("database", d.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Drop database failed")
		return err
	}

	d.logger.Warn().
		Str("operation", "drop_database").
		Str("database", d.name).
		Dur("duration_ms", duration).
		Msg("Database dropped")

	return nil
}

// CreateCollection creates a new collection (T075)
func (d *databaseWrapper) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	err := d.database.CreateCollection(ctx, name, opts...)

	duration := time.Since(startTime)

	// Structured logging
	if err != nil {
		d.logger.Error().
			Str("operation", "create_collection").
			Str("database", d.name).
			Str("collection", name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("Create collection failed")
		return err
	}

	d.logger.Info().
		Str("operation", "create_collection").
		Str("database", d.name).
		Str("collection", name).
		Dur("duration_ms", duration).
		Msg("Collection created")

	return nil
}

// ListCollectionNames returns all collection names (T076)
func (d *databaseWrapper) ListCollectionNames(ctx context.Context, filter interface{}) ([]string, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	startTime := time.Now()

	names, err := d.database.ListCollectionNames(ctx, filter)

	duration := time.Since(startTime)

	// Structured logging
	if err != nil {
		d.logger.Error().
			Str("operation", "list_collections").
			Str("database", d.name).
			Dur("duration_ms", duration).
			Err(err).
			Msg("List collections failed")
		return nil, err
	}

	d.logger.Debug().
		Str("operation", "list_collections").
		Str("database", d.name).
		Int("collection_count", len(names)).
		Dur("duration_ms", duration).
		Msg("Collections listed")

	return names, nil
}

// Collection returns a Collection interface for the named collection
func (d *databaseWrapper) Collection(name string) Collection {
	mongoCollection := d.database.Collection(name)
	return newCollection(mongoCollection, d.operationTimeout, d.logger)
}
