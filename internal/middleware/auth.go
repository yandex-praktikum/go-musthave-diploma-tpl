package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"net/http"
)

// Определяем собственный тип для ключей контекста
type contextKey string

const (
	LoginContentKey    contextKey = "Login"
	PasswordContentKey contextKey = "Password"
	AccessTokenKey     contextKey = "accessToken"
	AuthKey            contextKey = "auth"
)

type AuthMiddleware struct {
	authService *auth.ServiceAuth
}

func NewAuthMiddleware(authService *auth.ServiceAuth) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

//func (a *AuthMiddleware) AuthMiddleware(h http.Handler) http.Handler {
//	fn := func(w http.ResponseWriter, r *http.Request) {
//		authHeader := r.Header.Get("Authorization")
//		var accessToken string
//		if authHeader != "" {
//			accessToken = authHeader
//		} else {
//			//читаем токен из кук
//			cookie, err := r.Cookie(string(AccessTokenKey))
//			if err == nil && cookie.Value != "" {
//				accessToken = cookie.Value
//			} else {
//				accessToken = ""
//			}
//		}
//
//		userLogin, err := a.authService.VerifyUser(accessToken)
//		if err != nil {
//
//			// создаем токен
//			userLogin = uuid.New().String()
//			token, err := a.authService.CreatTokenForUser(userLogin)
//			if err != nil {
//				http.Error(w, `{"error":"Failed to generate auth token"}`, http.StatusInternalServerError)
//				return
//			}
//
//			//создаем куку
//			http.SetCookie(w, &http.Cookie{
//				Name:     string(AccessTokenKey),
//				Value:    token,
//				HttpOnly: true,
//				Path:     "/",
//			})
//
//			// Устанавливаем заголовок Authorization
//			w.Header().Set("Authorization", token)
//		}
//		ctxWithUser := context.WithValue(r.Context(), AuthKey, userLogin)
//		h.ServeHTTP(w, r.WithContext(ctxWithUser))
//	}
//
//	return http.HandlerFunc(fn)
//}

func (a *AuthMiddleware) ValidAuth(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie(string(AccessTokenKey))
		if err != nil {
			apiError, _ := json.Marshal(customerrors.APIError{Message: "access token not found"})
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(apiError)
			return
		}

		userLogin, err := a.authService.VerifyUser(accessToken.Value)
		if err != nil {
			apiError, _ := json.Marshal(customerrors.APIError{Message: err.Error()})
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(apiError)
			return
		}

		ctxWithUser := context.WithValue(r.Context(), LoginContentKey, userLogin)
		fmt.Printf("Пользователь %s - авторизован", userLogin)
		h.ServeHTTP(w, r.WithContext(ctxWithUser))

	}

	return http.HandlerFunc(fn)
}
