package integration

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/air-go/internal/db"
)

// TestConnectionPoolLoad tests connection pool under concurrent load (T099)
// Verifies pool handles 10-20 concurrent operations as per FR-006
func TestConnectionPoolLoad(t *testing.T) {
	ctx := context.Background()

	// Start test container
	_, uri, cleanup, err := StartTestContainerWithURI(ctx)
	require.NoError(t, err)
	defer cleanup()

	// Create client with pool size 10
	logger := zerolog.Nop()
	config := &db.DBConfig{
		URI:              uri,
		Database:         "load_test_db",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	require.NoError(t, err)

	connectCtx, connectCancel := context.WithTimeout(ctx, 30*time.Second)
	err = client.Connect(connectCtx)
	connectCancel()
	require.NoError(t, err)

	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		client.Close()
	}()

	collection := client.Collection("load_test")

	t.Run("Concurrent_10_Operations", func(t *testing.T) {
		testConcurrentOperations(t, collection, 10)
	})

	t.Run("Concurrent_20_Operations", func(t *testing.T) {
		testConcurrentOperations(t, collection, 20)
	})

	t.Run("Sustained_Load_100_Operations", func(t *testing.T) {
		testSustainedLoad(t, collection, 100, 10)
	})
}

// testConcurrentOperations tests N concurrent database operations
func testConcurrentOperations(t *testing.T, collection db.Collection, concurrency int) {
	ctx := context.Background()

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var errorCount atomic.Int64

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Perform insert operation
			doc := bson.M{
				"worker_id": id,
				"timestamp": time.Now(),
				"data":      fmt.Sprintf("worker %d data", id),
			}

			_, err := collection.InsertOne(ctx, doc)
			if err != nil {
				errorCount.Add(1)
				t.Logf("Worker %d insert failed: %v", id, err)
				return
			}

			successCount.Add(1)
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()
	duration := time.Since(startTime)

	// Verify all operations succeeded
	assert.Equal(t, int64(concurrency), successCount.Load(),
		"All %d concurrent operations should succeed", concurrency)
	assert.Equal(t, int64(0), errorCount.Load(),
		"No operations should fail under normal load")

	// Log performance metrics
	t.Logf("Completed %d concurrent operations in %v (avg: %v per op)",
		concurrency, duration, duration/time.Duration(concurrency))

	// Verify reasonable performance (operations should complete within a few seconds)
	assert.Less(t, duration, 10*time.Second,
		"Concurrent operations should complete within 10 seconds")
}

// testSustainedLoad tests sustained concurrent load over time
func testSustainedLoad(t *testing.T, collection db.Collection, totalOperations, concurrency int) {
	ctx := context.Background()

	var successCount atomic.Int64
	var errorCount atomic.Int64

	// Channel to control operation rate
	operationChan := make(chan int, totalOperations)
	for i := 0; i < totalOperations; i++ {
		operationChan <- i
	}
	close(operationChan)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Start worker goroutines
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for opID := range operationChan {
				// Insert document
				doc := bson.M{
					"worker_id":    workerID,
					"operation_id": opID,
					"timestamp":    time.Now(),
				}

				_, err := collection.InsertOne(ctx, doc)
				if err != nil {
					errorCount.Add(1)
					continue
				}

				// Query the document
				var result bson.M
				err = collection.FindOne(ctx, bson.M{"operation_id": opID}).Decode(&result)
				if err != nil {
					errorCount.Add(1)
					continue
				}

				successCount.Add(1)
			}
		}(w)
	}

	// Wait for all workers to complete
	wg.Wait()
	duration := time.Since(startTime)

	// Verify performance
	successOps := successCount.Load()
	errorOps := errorCount.Load()

	t.Logf("Sustained load: %d operations, %d workers, duration: %v",
		totalOperations, concurrency, duration)
	t.Logf("Success: %d, Errors: %d, Success rate: %.2f%%",
		successOps, errorOps, float64(successOps)/float64(totalOperations)*100)

	// Verify acceptable success rate (at least 95%)
	successRate := float64(successOps) / float64(totalOperations)
	assert.GreaterOrEqual(t, successRate, 0.95,
		"At least 95%% of operations should succeed under sustained load")

	// Calculate throughput
	throughput := float64(successOps) / duration.Seconds()
	t.Logf("Throughput: %.2f operations/second", throughput)

	// Verify reasonable throughput (at least 10 ops/sec for development)
	assert.GreaterOrEqual(t, throughput, 10.0,
		"Should achieve at least 10 operations per second")
}

// TestConnectionPoolExhaustion tests pool behavior when exhausted (T099)
func TestConnectionPoolExhaustion(t *testing.T) {
	ctx := context.Background()

	_, uri, cleanup, err := StartTestContainerWithURI(ctx)
	require.NoError(t, err)
	defer cleanup()

	// Create client with small pool size for testing
	logger := zerolog.Nop()
	config := &db.DBConfig{
		URI:              uri,
		Database:         "pool_test_db",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      2,
		MaxPoolSize:      5, // Small pool for testing exhaustion
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	require.NoError(t, err)

	connectCtx, connectCancel := context.WithTimeout(ctx, 30*time.Second)
	err = client.Connect(connectCtx)
	connectCancel()
	require.NoError(t, err)

	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		client.Close()
	}()

	collection := client.Collection("exhaust_test")

	// Launch more operations than pool size
	var wg sync.WaitGroup
	var successCount atomic.Int64
	var timeoutCount atomic.Int64

	concurrency := 10 // More than MaxPoolSize (5)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Use shorter timeout to detect pool exhaustion faster
			opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			doc := bson.M{"id": id, "data": "test"}
			_, err := collection.InsertOne(opCtx, doc)

			if err != nil {
				if opCtx.Err() == context.DeadlineExceeded {
					timeoutCount.Add(1)
				}
				return
			}

			successCount.Add(1)
		}(i)
	}

	wg.Wait()

	// With pool size 5, operations should eventually complete
	// Some might timeout if they wait too long, but most should succeed
	t.Logf("Pool exhaustion test: %d success, %d timeouts (pool size: 5)",
		successCount.Load(), timeoutCount.Load())

	// Verify that at least some operations succeeded
	assert.Greater(t, successCount.Load(), int64(0),
		"Some operations should succeed even with pool pressure")
}
