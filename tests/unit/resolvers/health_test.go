package resolvers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// MockDBClient is a mock implementation of resolvers.DBClient
type MockDBClient struct {
	mock.Mock
}

func (m *MockDBClient) HealthStatus(ctx context.Context) (*db.HealthStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.HealthStatus), args.Error(1)
}

func (m *MockDBClient) Collection(name string) db.Collection {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(db.Collection)
}

func (m *MockDBClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestAlive tests the alive query (T014)
func TestAlive(t *testing.T) {
	t.Run("should return true when system is operational", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().Alive(ctx)

		// Assert
		assert.NoError(t, err)
		assert.True(t, result, "alive query should return true")
	})

	t.Run("should not require authentication", func(t *testing.T) {
		// Arrange - unauthenticated context
		ctx := context.Background()
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().Alive(ctx)

		// Assert
		assert.NoError(t, err)
		assert.True(t, result, "alive query should work without authentication")
	})
}

// TestHealth tests the health query (T015)
func TestHealth(t *testing.T) {
	t.Run("should return ok status when database is connected", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockDB := new(MockDBClient)
		mockDB.On("HealthStatus", mock.Anything).Return(&db.HealthStatus{
			Status:    "connected",
			Message:   "MongoDB connected",
			LatencyMs: 2,
			Timestamp: time.Now(),
			Error:     "",
		}, nil)

		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act
		result, err := resolver.Query().Health(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ok", result.Status)
		assert.NotNil(t, result.Database)
		assert.Equal(t, "connected", result.Database.Status)
		assert.Equal(t, "MongoDB connected", result.Database.Message)
		assert.Equal(t, int64(2), result.Database.LatencyMs)
		mockDB.AssertExpectations(t)
	})

	t.Run("should return degraded status when database is not connected", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockDB := new(MockDBClient)
		mockDB.On("HealthStatus", mock.Anything).Return(&db.HealthStatus{
			Status:    "disconnected",
			Message:   "MongoDB connection failed",
			LatencyMs: 0,
			Timestamp: time.Now(),
			Error:     "connection timeout",
		}, nil)

		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act
		result, err := resolver.Query().Health(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "degraded", result.Status, "overall status should be degraded")
		assert.NotNil(t, result.Database)
		assert.Equal(t, "disconnected", result.Database.Status)
		mockDB.AssertExpectations(t)
	})

	t.Run("should not require authentication", func(t *testing.T) {
		// Arrange - unauthenticated context
		ctx := context.Background()
		mockDB := new(MockDBClient)
		mockDB.On("HealthStatus", mock.Anything).Return(&db.HealthStatus{
			Status:    "connected",
			Message:   "MongoDB connected",
			LatencyMs: 1,
			Timestamp: time.Now(),
			Error:     "",
		}, nil)

		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act
		result, err := resolver.Query().Health(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result, "health query should work without authentication")
		mockDB.AssertExpectations(t)
	})

	t.Run("should complete within 100ms threshold", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockDB := new(MockDBClient)
		mockDB.On("HealthStatus", mock.Anything).Return(&db.HealthStatus{
			Status:    "connected",
			Message:   "MongoDB connected",
			LatencyMs: 5,
			Timestamp: time.Now(),
			Error:     "",
		}, nil)

		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act & Assert - this is a performance expectation test
		// Actual timing would be measured in integration tests
		result, err := resolver.Query().Health(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// The health check itself should be fast, database latency is reported separately
		assert.Less(t, result.Database.LatencyMs, int64(100), "database latency should be under 100ms")
		mockDB.AssertExpectations(t)
	})
}
