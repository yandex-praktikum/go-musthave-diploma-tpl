package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenExpiration            = time.Hour * 3
	ClaimsKey       ContextKey = "claims"
)

// Middleware для проверки JWT
func AuthMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				if err == http.ErrNoCookie {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			claims := &common.Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(key), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Вспомогательная функция для получения данных из контекста
func GetClaims(ctx context.Context) *common.Claims {
	if claims, ok := ctx.Value(ClaimsKey).(*common.Claims); ok {
		return claims
	}
	return nil
}
