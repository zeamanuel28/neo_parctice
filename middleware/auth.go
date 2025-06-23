package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RoleKey   contextKey = "role" // ✅ Add role key
)

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Expecting format: Bearer <token>
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract user_id and role
		userID := claims["user_id"]
		role := claims["role"]

		// Add both to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, RoleKey, role)
	})
}

// Helper to get user ID from context in protected handlers
func GetUserIDFromContext(r *http.Request) string {
	id := r.Context().Value(UserIDKey)
	if id != nil {
		return fmt.Sprintf("%v", id)
	}
	return ""
}

// GetUserRoleFromContext extracts the user's role from the request context
func GetUserRoleFromContext(r *http.Request) string {
	role := r.Context().Value(RoleKey)
	if role != nil {
		return fmt.Sprintf("%v", role)
	}
	return ""
}
