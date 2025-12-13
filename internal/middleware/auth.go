package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/auth"
)

// AuthMiddleware проверяет JWT токен в запросе
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем публичные эндпоинты
		if r.URL.Path == "/api/user/register" ||
			r.URL.Path == "/api/user/register/" ||
			r.URL.Path == "/api/user/login" ||
			r.URL.Path == "/api/user/login/" {
			next.ServeHTTP(w, r)
			return
		}

		// Пробуем получить токен из cookie
		token := ""
		if cookie, err := r.Cookie("auth_token"); err == nil {
			token = cookie.Value
		}

		// Если нет в cookie, пробуем из заголовка Authorization
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Если токен не найден
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		claims, err := auth.VerifyToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Добавляем user_id в контекст
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_login", claims.Login)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetAuthToken устанавливает токен в cookie и заголовок
func SetAuthToken(w http.ResponseWriter, token string) {
	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Установите true для HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Также устанавливаем в заголовок Authorization
	w.Header().Set("Authorization", "Bearer "+token)
}
