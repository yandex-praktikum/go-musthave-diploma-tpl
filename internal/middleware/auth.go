package middleware

import (
	"context"
	"net/http"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
)

var ctxKey = "UserLogin"

func CheckAuthorization(h http.Handler) http.Handler {
	login := func(res http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		claims, ok := storage.VerifyToken(token)
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		var userData storage.UserData
		userData.Login = claims["sub"].(string)
		ok, err := storage.CheckUserExists(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, userData.Login))
		h.ServeHTTP(res, req)
	}
	return http.HandlerFunc(login)
}
