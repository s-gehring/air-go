package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for storing the user ID from JWT
	UserIDKey ContextKey = "user_id"
	// ClaimsKey is the context key for storing the full JWT claims
	ClaimsKey ContextKey = "claims"
)

// AuthMiddleware validates JWT tokens and adds user information to the request context
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn().Msg("Missing Authorization header")
				http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Check for "Bearer " prefix
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				log.Warn().Msg("Invalid Authorization header format")
				http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				log.Warn().Err(err).Msg("Token validation failed")
				http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				log.Warn().Msg("Invalid token")
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Warn().Msg("Failed to parse token claims")
				http.Error(w, "Unauthorized: Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Extract user ID from claims (sub claim is standard)
			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				log.Warn().Msg("Missing or invalid user ID in token claims")
				http.Error(w, "Unauthorized: Missing user ID in token", http.StatusUnauthorized)
				return
			}

			// Add user ID and full claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			log.Debug().Str("user_id", userID).Msg("User authenticated successfully")

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetClaims extracts the JWT claims from the request context
func GetClaims(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(jwt.MapClaims)
	return claims, ok
}
