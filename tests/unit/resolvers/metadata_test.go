package resolvers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"github.com/yourusername/air-go/tests/testutil"
)

// TestErrorCodeMetadataGet tests the errorCodeMetadataGet query (T016)
func TestErrorCodeMetadataGet(t *testing.T) {
	t.Run("should return all error code metadata", func(t *testing.T) {
		// Arrange
		ctx := testutil.WithUserContext(context.Background(), "test-user", "user@test.com")
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().ErrorCodeMetadataGet(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, []*generated.ErrorCodeMetadata{}, result)
		// The actual metadata would be loaded from database/config
		// For now, we expect an empty list or populated list
	})

	t.Run("should require authentication", func(t *testing.T) {
		// Arrange - unauthenticated context
		ctx := context.Background()
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().ErrorCodeMetadataGet(ctx)

		// Assert
		assert.Error(t, err, "should require authentication")
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Authentication required")
	})

	t.Run("should work for authenticated users", func(t *testing.T) {
		// Arrange
		ctx := testutil.WithAuthContext(context.Background(), &testutil.MockUserClaims{
			UserID: "test-user",
			Email:  "user@test.com",
			Roles:  []string{"USER"},
		})
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().ErrorCodeMetadataGet(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestInconsistencyMetadataGet tests the inconsistencyMetadataGet query (T017)
func TestInconsistencyMetadataGet(t *testing.T) {
	t.Run("should return all inconsistency metadata", func(t *testing.T) {
		// Arrange
		ctx := testutil.WithUserContext(context.Background(), "test-user", "user@test.com")
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().InconsistencyMetadataGet(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, []*generated.InconsistencyMetadata{}, result)
	})

	t.Run("should require authentication", func(t *testing.T) {
		// Arrange - unauthenticated context
		ctx := context.Background()
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().InconsistencyMetadataGet(ctx)

		// Assert
		assert.Error(t, err, "should require authentication")
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Authentication required")
	})

	t.Run("should work for authenticated advisors", func(t *testing.T) {
		// Arrange
		ctx := testutil.WithAdvisorContext(context.Background())
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().InconsistencyMetadataGet(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestDocumentMetadataGet tests the documentMetadataGet query (T018)
func TestDocumentMetadataGet(t *testing.T) {
	t.Run("should return all business document metadata", func(t *testing.T) {
		// Arrange
		ctx := testutil.WithUserContext(context.Background(), "test-user", "user@test.com")
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().DocumentMetadataGet(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, []*generated.BizDocMetadata{}, result)
	})

	t.Run("should require authentication", func(t *testing.T) {
		// Arrange - unauthenticated context
		ctx := context.Background()
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().DocumentMetadataGet(ctx)

		// Assert
		assert.Error(t, err, "should require authentication")
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Authentication required")
	})

	t.Run("should work for admin users", func(t *testing.T) {
		// Arrange
		ctx := testutil.WithAdminContext(context.Background())
		resolver := &resolvers.Resolver{}

		// Act
		result, err := resolver.Query().DocumentMetadataGet(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
