package routers

import (
	"context"
	"net/http"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/auth"
)

type contextKey string

var userIDKey = contextKey("userID")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		userID, err := auth.ParseJWT(cookie.Value)
		if err != nil {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
