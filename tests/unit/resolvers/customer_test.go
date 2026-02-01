package resolvers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MockCollection is a mock implementation of db.Collection interface
type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents)
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollection) UpdateMany(ctx context.Context, filter, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollection) DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollection) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCollection) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, pipeline, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

// MockDBClient is a mock implementation of resolvers.DBClient interface
type MockCustomerDBClient struct {
	mock.Mock
}

func (m *MockCustomerDBClient) Collection(name string) db.Collection {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(db.Collection)
}

func (m *MockCustomerDBClient) HealthStatus(ctx context.Context) (*db.HealthStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.HealthStatus), args.Error(1)
}

func (m *MockCustomerDBClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestCustomerGet_InvalidUUID tests UUID validation (T008)
func TestCustomerGet_InvalidUUID(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantError  bool
	}{
		{
			name:       "empty UUID",
			identifier: "",
			wantError:  true,
		},
		{
			name:       "malformed UUID - not-a-uuid",
			identifier: "not-a-uuid",
			wantError:  true,
		},
		{
			name:       "malformed UUID - 123",
			identifier: "123",
			wantError:  true,
		},
		{
			name:       "incomplete UUID",
			identifier: "550e8400-e29b-41d4-a716",
			wantError:  true,
		},
		{
			name:       "UUID with extra characters",
			identifier: "550e8400-e29b-41d4-a716-446655440000-extra",
			wantError:  true,
		},
		{
			name:       "UUID with spaces",
			identifier: "550e8400 e29b 41d4 a716 446655440000",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			resolver := &resolvers.Resolver{}

			// Act
			customer, err := resolver.Query().CustomerGet(ctx, tt.identifier)

			// Assert
			if tt.wantError {
				assert.Error(t, err, "Expected error for invalid UUID: %s", tt.identifier)
				assert.Nil(t, customer, "Customer should be nil for invalid UUID")
				assert.Contains(t, err.Error(), "invalid", "Error message should mention 'invalid'")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCustomerGet_NotFound tests null return when customer not found (T009)
func TestCustomerGet_NotFound(t *testing.T) {
	t.Run("should return nil for non-existent customer", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		identifier := "550e8400-e29b-41d4-a716-446655440000"
		
		mockDB := new(MockCustomerDBClient)
		mockColl := new(MockCollection)
		
		// Mock FindOne to return ErrNoDocuments
		singleResult := &mongo.SingleResult{}
		// Note: In reality, this would be set up to return ErrNoDocuments
		mockColl.On("FindOne", ctx, mock.MatchedBy(func(filter interface{}) bool {
			m, ok := filter.(bson.M)
			if !ok {
				return false
			}
			return m["identifier"] == identifier
		})).Return(singleResult)
		
		mockDB.On("Collection", "customers").Return(mockColl)
		
		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act
		customer, err := resolver.Query().CustomerGet(ctx, identifier)

		// Assert
		assert.NoError(t, err, "Should not return error for non-existent customer")
		assert.Nil(t, customer, "Should return nil for non-existent customer")
		mockDB.AssertExpectations(t)
		mockColl.AssertExpectations(t)
	})
}

// TestCustomerGet_Deleted tests null return when customer is deleted (T010)
func TestCustomerGet_Deleted(t *testing.T) {
	t.Run("should return nil for deleted customer", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		identifier := "550e8400-e29b-41d4-a716-446655440000"
		
		mockDB := new(MockCustomerDBClient)
		mockColl := new(MockCollection)
		
		// Mock FindOne to check filter excludes deleted customers
		singleResult := &mongo.SingleResult{}
		mockColl.On("FindOne", ctx, mock.MatchedBy(func(filter interface{}) bool {
			m, ok := filter.(bson.M)
			if !ok {
				return false
			}
			// Verify filter includes deletion status exclusion
			_, exists := m["status.deletion"]
			return exists && m["identifier"] == identifier
		})).Return(singleResult)
		
		mockDB.On("Collection", "customers").Return(mockColl)
		
		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act
		customer, err := resolver.Query().CustomerGet(ctx, identifier)

		// Assert
		assert.NoError(t, err, "Should not return error for deleted customer")
		assert.Nil(t, customer, "Should return nil for deleted customer")
		mockDB.AssertExpectations(t)
		mockColl.AssertExpectations(t)
	})
}
