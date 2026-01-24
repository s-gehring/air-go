package db

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client encapsulates MongoDB client connection with lifecycle management
type Client struct {
	// MongoDB driver client
	mongoClient *mongo.Client

	// Configuration
	config   *DBConfig
	database *mongo.Database

	// State tracking
	connected atomic.Bool
	lastPing  time.Time
	mu        sync.RWMutex

	// Health cache
	healthCache *healthCache
	healthMu    sync.RWMutex

	// Context for lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Logger
	logger zerolog.Logger
}

// NewClient creates a new MongoDB client instance
func NewClient(config *DBConfig, logger zerolog.Logger) (*Client, error) {
	if config == nil {
		return nil, ErrInvalidConfiguration
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		healthCache: &healthCache{
			expiresAt: time.Now(),
		},
	}

	return client, nil
}

// IsConnected returns the current connection state (thread-safe, cached)
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

// Database returns the database instance for admin operations (T073)
// Returns a Database interface for admin-level operations
func (c *Client) Database() Database {
	if c.database == nil {
		return nil
	}
	return newDatabase(c.database, c.config.OperationTimeout, c.logger)
}

// Connect establishes connection to MongoDB with automatic retry logic
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already connected
	if c.connected.Load() {
		return ErrAlreadyConnected
	}

	// Create client options with connection pool settings
	clientOptions := options.Client().
		ApplyURI(c.config.URI).
		SetMinPoolSize(c.config.MinPoolSize).
		SetMaxPoolSize(c.config.MaxPoolSize).
		SetMaxConnIdleTime(c.config.MaxConnIdleTime).
		SetServerSelectionTimeout(c.config.ConnectTimeout)

	retryState := &RetryState{
		Attempt: 0,
	}

	startTime := time.Now()

	// Retry loop with exponential backoff
	for attempt := 1; attempt <= c.config.MaxRetryAttempts; attempt++ {
		retryState.Attempt = attempt

		c.logger.Info().
			Str("event_type", "mongodb_connection_attempt").
			Str("host", c.config.URI).
			Str("database", c.config.Database).
			Int("attempt", attempt).
			Msg("Connecting to MongoDB")

		// Create MongoDB client
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			retryState.LastError = err
			retryState.TotalDuration = time.Since(startTime)

			// If this is the last attempt or context is cancelled, fail
			if !ShouldRetry(attempt, c.config.MaxRetryAttempts) || ctx.Err() != nil {
				c.logger.Error().
					Str("event_type", "mongodb_connection_error").
					Str("host", c.config.URI).
					Str("database", c.config.Database).
					Err(err).
					Msg("Failed to connect to MongoDB")
				return ErrConnectionTimeout
			}

			// Calculate delay and wait before retry
			delay := CalculateDelay(attempt, c.config.RetryBaseDelay, c.config.RetryMaxDelay)
			LogRetryAttempt(c.logger, retryState, delay)

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Ping to verify connection
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = client.Ping(pingCtx, nil)
		cancel()

		if err != nil {
			client.Disconnect(context.Background())
			retryState.LastError = err
			retryState.TotalDuration = time.Since(startTime)

			if !ShouldRetry(attempt, c.config.MaxRetryAttempts) || ctx.Err() != nil {
				c.logger.Error().
					Str("event_type", "mongodb_connection_error").
					Str("host", c.config.URI).
					Err(err).
					Msg("Failed to ping MongoDB")
				return ErrConnectionTimeout
			}

			delay := CalculateDelay(attempt, c.config.RetryBaseDelay, c.config.RetryMaxDelay)
			LogRetryAttempt(c.logger, retryState, delay)

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Connection successful
		c.mongoClient = client
		c.database = client.Database(c.config.Database)
		c.connected.Store(true)
		c.lastPing = time.Now()

		latency := time.Since(startTime)

		c.logger.Info().
			Str("event_type", "mongodb_connection_established").
			Str("host", c.config.URI).
			Str("database", c.config.Database).
			Int64("latency_ms", latency.Milliseconds()).
			Uint64("pool_size", c.config.MaxPoolSize).
			Msg("MongoDB connected")

		return nil
	}

	return ErrConnectionTimeout
}

// Disconnect gracefully closes MongoDB connection and cleanup resources
func (c *Client) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected.Load() {
		return nil // Already disconnected, no error
	}

	if c.mongoClient == nil {
		return nil
	}

	startTime := time.Now()

	err := c.mongoClient.Disconnect(ctx)
	if err != nil {
		c.logger.Error().
			Str("event_type", "mongodb_disconnect_error").
			Err(err).
			Msg("Error during MongoDB disconnection")
		// Don't return error, mark as disconnected anyway
	}

	c.connected.Store(false)
	c.mongoClient = nil
	c.database = nil

	duration := time.Since(startTime)

	c.logger.Info().
		Str("event_type", "mongodb_disconnected").
		Str("host", c.config.URI).
		Str("database", c.config.Database).
		Int64("duration_ms", duration.Milliseconds()).
		Msg("MongoDB disconnected")

	return nil
}

// Ping verifies MongoDB connectivity with lightweight operation
func (c *Client) Ping(ctx context.Context) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	if c.mongoClient == nil {
		return ErrNotConnected
	}

	err := c.mongoClient.Ping(ctx, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.lastPing = time.Now()
	c.mu.Unlock()

	return nil
}

// HealthStatus returns comprehensive health check information with caching
func (c *Client) HealthStatus(ctx context.Context) (*HealthStatus, error) {
	// Check cache (5-second TTL)
	c.healthMu.RLock()
	if c.healthCache.status != nil && time.Now().Before(c.healthCache.expiresAt) {
		cached := c.healthCache.status
		c.healthMu.RUnlock()
		return cached, nil
	}
	c.healthMu.RUnlock()

	// Perform health check
	status := &HealthStatus{
		Timestamp: time.Now(),
	}

	if !c.connected.Load() {
		status.Status = "disconnected"
		status.Message = "MongoDB not connected"
		status.LatencyMs = 0
	} else {
		startTime := time.Now()
		err := c.Ping(ctx)
		latency := time.Since(startTime)

		if err != nil {
			status.Status = "error"
			status.Message = "Ping failed"
			status.LatencyMs = latency.Milliseconds()
			status.Error = err.Error()
		} else {
			status.Status = "connected"
			status.Message = "MongoDB connected"
			status.LatencyMs = latency.Milliseconds()
		}
	}

	// Update cache
	c.healthMu.Lock()
	c.healthCache.status = status
	c.healthCache.expiresAt = time.Now().Add(5 * time.Second)
	c.healthMu.Unlock()

	return status, nil
}

// Collection returns a collection accessor for database operations (T059)
// Returns a Collection interface with timeout enforcement and structured logging
func (c *Client) Collection(name string) Collection {
	if c.database == nil {
		panic("database not initialized: call Connect() first")
	}

	mongoCollection := c.database.Collection(name)
	return newCollection(mongoCollection, c.config.OperationTimeout, c.logger)
}

// Close gracefully shuts down the client and cancels the context
func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}
