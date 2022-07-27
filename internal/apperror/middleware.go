package apperror

import (
	"context"
	"errors"
	"github.com/botaevg/gophermart/internal/service"
	"github.com/dgrijalva/jwt-go/v4"
	"log"
	"net/http"
)

type AuthMiddleware struct {
	secretkey string
}

func NewAuthMiddleware(key string) *AuthMiddleware {
	return &AuthMiddleware{
		secretkey: key,
	}
}

type UserID string

func (a AuthMiddleware) AuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openURL := []string{"/api/user/register", "/api/user/login"}
		path := r.URL.Path
		for _, v := range openURL {
			if v == path {
				next.ServeHTTP(w, r)
				return
			}
		}
		ctoken, err := r.Cookie("Bearer")
		if err != nil {
			http.Error(w, errors.New("unauthorized").Error(), http.StatusUnauthorized)
			return
		}

		tokenClaims := &service.Claims{}
		token, err := jwt.ParseWithClaims(ctoken.Value, tokenClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.secretkey), nil
		})
		if err != nil {
			http.Error(w, errors.New("token error").Error(), http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, errors.New("token disabled").Error(), http.StatusUnauthorized)
			return
		}
		log.Print("userID " + string(tokenClaims.UserID))
		ctx := context.WithValue(r.Context(), UserID("username"), tokenClaims.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}
