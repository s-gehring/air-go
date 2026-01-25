package resolvers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockMongoCollection is a mock implementation of MongoDB collection
type MockMongoCollection struct {
	mock.Mock
}

func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

// MockDBClient is a mock implementation of DBClient
type MockDBClient struct {
	mock.Mock
}

func (m *MockDBClient) Collection(name string) interface{} {
	args := m.Called(name)
	return args.Get(0)
}

func (m *MockDBClient) HealthStatus(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
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
		
		mockDB := new(MockDBClient)
		mockColl := new(MockMongoCollection)
		
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
		
		mockDB := new(MockDBClient)
		mockColl := new(MockMongoCollection)
		
		// Mock FindOne to check filter excludes deleted customers
		singleResult := &mongo.SingleResult{}
		mockColl.On("FindOne", ctx, mock.MatchedBy(func(filter interface{}) bool {
			m, ok := filter.(bson.M)
			if !ok {
				return false
			}
			// Verify filter includes deletion status exclusion
			deletionFilter, exists := m["status.deletion"]
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

// TestCustomerGet_Success tests successful customer retrieval (T011)
func TestCustomerGet_Success(t *testing.T) {
	t.Run("should return Customer object for valid non-deleted customer", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		identifier := "550e8400-e29b-41d4-a716-446655440000"
		firstName := "John"
		lastName := "Doe"
		
		expectedCustomer := &generated.Customer{
			Identifier:      identifier,
			FirstName:       &firstName,
			LastName:        &lastName,
			ActionIndicator: generated.ActionIndicatorNone,
		}
		
		mockDB := new(MockDBClient)
		mockColl := new(MockMongoCollection)
		
		// Mock FindOne to return a customer
		singleResult := &mongo.SingleResult{}
		// Note: In reality, we'd need to mock the Decode method on SingleResult
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
		// This test will FAIL until implementation is complete
		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, identifier, customer.Identifier)
		mockDB.AssertExpectations(t)
		mockColl.AssertExpectations(t)
	})
}

// TestCustomerGet_DatabaseError tests database error handling (T012)
func TestCustomerGet_DatabaseError(t *testing.T) {
	t.Run("should return DATABASE_ERROR for connection failures", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		identifier := "550e8400-e29b-41d4-a716-446655440000"
		
		mockDB := new(MockDBClient)
		mockColl := new(MockMongoCollection)
		
		// Mock FindOne to return an error
		dbError := errors.New("connection timeout")
		mockColl.On("FindOne", ctx, mock.Anything).Return((*mongo.SingleResult)(nil)).Maybe()
		
		mockDB.On("Collection", "customers").Return(mockColl)
		
		resolver := &resolvers.Resolver{
			DBClient: mockDB,
		}

		// Act
		customer, err := resolver.Query().CustomerGet(ctx, identifier)

		// Assert
		assert.Error(t, err, "Should return error for database failure")
		assert.Nil(t, customer, "Customer should be nil on database error")
		assert.Contains(t, err.Error(), "database", "Error should indicate database issue")
		mockDB.AssertExpectations(t)
	})
}

// TestIsValidUUID tests UUID validation function (T007)
func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		uuid  string
		valid bool
	}{
		{
			name:  "valid lowercase UUID",
			uuid:  "550e8400-e29b-41d4-a716-446655440000",
			valid: true,
		},
		{
			name:  "valid uppercase UUID",
			uuid:  "550E8400-E29B-41D4-A716-446655440000",
			valid: true,
		},
		{
			name:  "valid mixed case UUID",
			uuid:  "550e8400-E29B-41d4-A716-446655440000",
			valid: true,
		},
		{
			name:  "invalid - empty string",
			uuid:  "",
			valid: false,
		},
		{
			name:  "invalid - not-a-uuid",
			uuid:  "not-a-uuid",
			valid: false,
		},
		{
			name:  "invalid - missing dashes",
			uuid:  "550e8400e29b41d4a716446655440000",
			valid: false,
		},
		{
			name:  "invalid - incomplete",
			uuid:  "550e8400-e29b-41d4-a716",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will FAIL until isValidUUID is implemented
			// For now, we're testing the expected behavior
			
			// Note: isValidUUID should be exported or we test through CustomerGet
			// Testing through CustomerGet for now
			ctx := context.Background()
			resolver := &resolvers.Resolver{}
			
			_, err := resolver.Query().CustomerGet(ctx, tt.uuid)
			
			if tt.valid {
				// Valid UUIDs should not fail validation (may fail DB lookup)
				// We're only checking that it doesn't fail with "invalid" error
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid", 
						"Valid UUID %s should not trigger invalid error", tt.uuid)
				}
			} else {
				// Invalid UUIDs should fail with validation error
				assert.Error(t, err, "Expected error for invalid UUID: %s", tt.uuid)
				assert.Contains(t, err.Error(), "invalid", 
					"Expected 'invalid' in error message for UUID: %s", tt.uuid)
			}
		})
	}
}
