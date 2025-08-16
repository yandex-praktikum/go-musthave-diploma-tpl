package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type contextLogin string

const (
	UserLoginKey contextLogin = "userID"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверяем куку auth_token
		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// 2. Если кука есть - используем её значение как login
		userID, err := strconv.Atoi(cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// 3. Добавляем login в контекст
		ctx := context.WithValue(r.Context(), UserLoginKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
