// internal/middleware/auth.go
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/b602op/go-musthave-diploma-tpl/internal/utils"
)

// contextKey — тип ключа контекста для хранения userID.
type contextKey string

const (
	// userIDKey — ключ контекста, в котором хранится идентификатор пользователя.
	userIDKey contextKey = "userID"
)

// AuthMiddleware структура для аутентификации
type AuthMiddleware struct {
	jwtUtils *utils.JWTUtils
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	log.Println("[INFO] AuthMiddleware initialized")

	return &AuthMiddleware{
		jwtUtils: utils.NewJWTUtils(secretKey),
	}
}

// GenerateToken генерирует JWT токен (обертка над utils)
func (m *AuthMiddleware) GenerateToken(userID int64) (string, error) {
	log.Printf("[DEBUG] Generating token for userID=%d", userID)
	token, err := m.jwtUtils.GenerateToken(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to generate token for userID=%d: %v", userID, err)
		return "", err
	}
	log.Printf("[DEBUG] Token generated successfully for userID=%d", userID)
	return token, nil
}

// Auth middleware проверяет JWT токен
func (m *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Auth: Checking request %s %s", r.Method, r.URL.Path)

		// Ищем токен в заголовке Authorization
		authHeader := r.Header.Get("Authorization")
		tokenString := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
				log.Printf("[DEBUG] Auth: Token found in Authorization header")
			} else {
				log.Printf("[WARN] Auth: Invalid Authorization header format: %s", authHeader)
			}
		}

		// Если в заголовке нет, ищем в cookie
		if tokenString == "" {
			cookie, err := r.Cookie("token")
			if err == nil {
				tokenString = cookie.Value
				log.Printf("[DEBUG] Auth: Token found in cookie")
			} else {
				log.Printf("[DEBUG] Auth: No token in cookie: %v", err)
			}
		}

		// Если токена нет - возвращаем 401
		if tokenString == "" {
			log.Printf("[WARN] Auth: No token found for %s %s", r.Method, r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем токен через utils
		claims, err := m.jwtUtils.ValidateToken(tokenString)
		if err != nil {
			log.Printf("[WARN] Auth: Invalid token for %s %s: %v", r.Method, r.URL.Path, err)

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("[DEBUG] Auth: Token valid for userID=%d", claims.UserID)

		// Сохраняем userID в контексте
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает userID из контекста
func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)

	if ok {
		log.Printf("[DEBUG] GetUserID: userID=%d found in context", userID)
	} else {
		log.Printf("[DEBUG] GetUserID: userID not found in context")
	}

	return userID, ok
}
