package testutil

import (
	"context"

	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// MockUserClaims represents test user claims
type MockUserClaims struct {
	UserID      string
	Email       string
	Roles       []string
	Permissions []string
}

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	userClaimsKey contextKey = "user_claims"
)

// WithAuthContext returns a context with user claims for testing
// Converts MockUserClaims to resolvers.UserClaims and stores in context using the resolver's helper
func WithAuthContext(ctx context.Context, claims *MockUserClaims) context.Context {
	// Convert to resolvers.UserClaims type
	resolverClaims := &resolvers.UserClaims{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}
	return resolvers.WithUserClaims(ctx, resolverClaims)
}

// WithAdminContext returns a context with admin user claims for testing
func WithAdminContext(ctx context.Context) context.Context {
	claims := &MockUserClaims{
		UserID:      "test-admin-id",
		Email:       "admin@test.com",
		Roles:       []string{"ADMIN"},
		Permissions: []string{"READ_ALL", "WRITE_ALL"},
	}
	return WithAuthContext(ctx, claims)
}

// WithUserContext returns a context with regular user claims for testing
func WithUserContext(ctx context.Context, userID, email string) context.Context {
	claims := &MockUserClaims{
		UserID:      userID,
		Email:       email,
		Roles:       []string{"USER"},
		Permissions: []string{"READ_OWN"},
	}
	return WithAuthContext(ctx, claims)
}

// WithAdvisorContext returns a context with financial advisor claims for testing
func WithAdvisorContext(ctx context.Context) context.Context {
	claims := &MockUserClaims{
		UserID:      "test-advisor-id",
		Email:       "advisor@test.com",
		Roles:       []string{"FINANCIAL_ADVISOR"},
		Permissions: []string{"READ_PORTFOLIOS", "READ_CUSTOMERS", "CREATE_PLANS"},
	}
	return WithAuthContext(ctx, claims)
}

// UnauthenticatedContext returns a context without any user claims
func UnauthenticatedContext(ctx context.Context) context.Context {
	return ctx
}
