package mw

import (
	resp "github.com/EestiChameleon/GOphermart/internal/app/router/responses"
	"github.com/EestiChameleon/GOphermart/internal/app/service"
	"github.com/EestiChameleon/GOphermart/internal/app/storage"
	"net/http"
)

func AuthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("gophermartID")
		if err != nil {
			resp.NoContent(w, http.StatusUnauthorized)
		}

		userID, err := service.JWTDecode(cookie.Value, "userID")
		if err != nil {
			resp.NoContent(w, http.StatusInternalServerError)
			return
		}

		storage.Pool.ID = int(userID.(float64))
		next.ServeHTTP(w, r)
	})
}
