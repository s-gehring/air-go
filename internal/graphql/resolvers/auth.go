package resolvers

import (
	"context"
	"errors"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	userClaimsKey contextKey = "user_claims"
)

// UserClaims represents the authenticated user's claims from JWT
type UserClaims struct {
	UserID      string
	Email       string
	Roles       []string
	Permissions []string
}

// requireAuth ensures the user is authenticated
// Returns the user claims if authentication is valid, or an error if not
func requireAuth(ctx context.Context) (*UserClaims, error) {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	if !ok || claims == nil {
		return nil, &QueryError{
			Message: "Authentication required",
			Code:    ErrCodeUnauthorized,
		}
	}
	return claims, nil
}

// requireAdmin ensures the user is authenticated and has admin role
// Returns the user claims if user is admin, or an error if not
func requireAdmin(ctx context.Context) (*UserClaims, error) {
	claims, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user has ADMIN role
	for _, role := range claims.Roles {
		if role == "ADMIN" || role == "ADMINISTRATOR" {
			return claims, nil
		}
	}

	return nil, &QueryError{
		Message: "Administrator privileges required",
		Code:    ErrCodeForbidden,
	}
}

// getUserClaims returns the user claims from context if present, nil otherwise
// This is a non-failing version that returns nil if no claims exist
func getUserClaims(ctx context.Context) *UserClaims {
	claims, _ := ctx.Value(userClaimsKey).(*UserClaims)
	return claims
}
