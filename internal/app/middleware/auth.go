package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
)

func (m *MiddleWare) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login := strings.HasPrefix(r.URL.Path, "/api/user/login")
		register := strings.HasPrefix(r.URL.Path, "/api/user/register")
		cookie, err := r.Cookie("go_session")

		if login || register {
			h.ServeHTTP(w, r)
			return
		} else if err != nil || !m.cookieValid(cookie.Value) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := m.GetUserIDFromCookie(cookie.Value)
		if err != nil {
			http.Error(w, "Wrong auth header", http.StatusUnauthorized)
		}

		ctx := context.WithValue(r.Context(), models.UserIDKey, userID)

		h.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (m *MiddleWare) cookieValid(cookie string) bool {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(cookie, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.conf.SecretKey), nil
		})
	if err != nil {
		return false
	}
	if !token.Valid {
		return false
	}

	userID := claims.UserID
	b := m.auth.FindUserByID(userID)

	return b
}

func (m *MiddleWare) GetUserIDFromCookie(cookieValue string) (string, error) {
	token, err := jwt.Parse(cookieValue, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.conf.SecretKey), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if userID, ok := claims["UserID"].(string); ok {
			return userID, nil
		}
	}

	return "", fmt.Errorf("invalid token")
}
