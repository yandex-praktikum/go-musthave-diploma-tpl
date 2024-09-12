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
		fmt.Printf("Пользователь %s - авторизован \n", userLogin)
		h.ServeHTTP(w, r.WithContext(ctxWithUser))

	}

	return http.HandlerFunc(fn)
}
