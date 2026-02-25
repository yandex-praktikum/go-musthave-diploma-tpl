// Package middleware содержит HTTP middleware для сервиса
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/anon-d/gophermarket/pkg/jwt"
)

const (
	// UserIDKey ключ для хранения user_id в контексте
	UserIDKey = "user_id"
)

// AuthMiddleware middleware для проверки JWT токена
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Пробуем получить токен из cookie
		if cookie, err := c.Cookie("token"); err == nil {
			token = cookie
		}

		// Пробуем получить токен из заголовка Authorization
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		tokenData, err := jwt.GetToken(token, []byte(jwtSecret))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := jwt.GetClaimsVerified(tokenData)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Получаем user_id из claims
		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Сохраняем user_id в контексте для передачи ниже
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID получает user_id из контекста
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	return userID.(string)
}
