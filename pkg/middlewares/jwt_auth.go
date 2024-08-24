package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/eac0de/gophermart/pkg/jwt"
)

type key string

const UserKey key = "User"

func GetJWTAuthMiddleware(tokenService *jwt.JWTTokenService, userStore UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/user/register" || r.URL.Path == "/api/user/login" || r.URL.Path == "/api/user/refresh_token" {
				next.ServeHTTP(w, r)
				return
			}
			authorization_header := r.Header.Get("Authorization")
			if authorization_header == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("You must register/login"))
				return
			}
			clear_token, ok := strings.CutPrefix(authorization_header, "Bearer ")
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Authorization header"))
				return
			}
			claims, err := tokenService.ValidateAccessToken(r.Context(), clear_token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid access token"))
				return
			}
			user, err := userStore.SelectUserByID(r.Context(), claims.UserID)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("User not found"))
				return
			}
			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromRequest(r *http.Request) *models.User {
	user, _ := r.Context().Value(UserKey).(*models.User)
	return user
}
