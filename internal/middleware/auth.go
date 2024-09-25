package middleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

const SecretKey = "sectet_key"
const UsernameKey = "username"

type Credentials struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		authHeader := r.Header.Get("Authorization")

		if err != nil && authHeader == "" {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		var tokenString string

		if cookie != nil {
			tokenString = cookie.Value
		} else {
			tokenString = authHeader
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "недействительный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UsernameKey, claims.Username)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
