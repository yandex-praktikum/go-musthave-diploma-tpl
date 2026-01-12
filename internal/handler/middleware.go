package handler

import (
	"context"
	"net/http"
	"strings"

	"gophermart/internal/service"
)

type contextKey string

const userIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtService *service.JWTService
}

func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := m.extractToken(r)
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) extractToken(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	return ""
}

func getUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
}

func getUserID(r *http.Request) int64 {
	userID, _ := getUserIDFromContext(r)
	return userID
}
