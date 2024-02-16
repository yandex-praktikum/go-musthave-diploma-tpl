package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
	"github.com/golang-jwt/jwt"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func Registration() http.Handler {
	register := func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		regData := RegisterRequest{}
		// читаем тело запроса
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		data := buf.Bytes()

		if err = json.Unmarshal(data, &regData); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		userData := storage.UserData{}
		userData.Login = regData.Login
		userData.Password = regData.Password
		result, err := storage.CheckUserExists(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if result {
			res.WriteHeader(http.StatusConflict)
			return
		}
		err = storage.CreateNewUser(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		payload := jwt.MapClaims{
			"sub": userData.Login,
			"exp": time.Now().Add(time.Hour * 72).Unix(),
		}

		// Создаем новый JWT-токен и подписываем его по алгоритму HS256
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

		authToken, err := token.SignedString(jwtSecretKey)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Add("Authorization", authToken)
		res.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(register)
}
