package auth

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// UserIDKey ключ для хранения userID в контексте
	UserIDKey ContextKey = "user_id"
)

// AuthMiddleware middleware для проверки аутентификации
type AuthMiddleware struct {
	tokenManager *TokenManager
}

// NewAuthMiddleware создаёт новый middleware аутентификации
func NewAuthMiddleware(tokenManager *TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenManager: tokenManager,
	}
}

// Middleware возвращает http.Handler который проверяет JWT токен
// Токен ожидается в заголовке Authorization в формате "Bearer <token>"
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
			return
		}

		// Ожидаем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Валидируем токен и извлекаем userID
		userID, err := am.tokenManager.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает userID из контекста запроса
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
