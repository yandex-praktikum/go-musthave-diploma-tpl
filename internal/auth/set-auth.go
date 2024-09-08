package auth

import (
	"context"
	"net/http"
	"time"
)

type contextLoginKey string

const loginKey contextLoginKey = "login"
const tokenExp = time.Hour * 24

func SetAuth(r *http.Request, w http.ResponseWriter, login string) {
	expTime := time.Now().Add(tokenExp)

	tokenString, err := GenerateJWT(login, expTime)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
	}

	ctx := context.WithValue(r.Context(), loginKey, login)
	r = r.WithContext(ctx)

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  expTime,
		Path:     "/",
		HttpOnly: true,
	})

	w.Header().Set("Authorization", "Bearer "+tokenString)
}
