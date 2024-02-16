package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
	"github.com/golang-jwt/jwt"
)

// Структура HTTP-запроса на вход в аккаунт
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Структура HTTP-ответа на вход в аккаунт
// В ответе содержится JWT-токен авторизованного пользователя
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

var jwtSecretKey = []byte(storage.SecretKey)

func LoginUser() http.Handler {
	login := func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		loginData := LoginRequest{}
		// читаем тело запроса
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		data := buf.Bytes()

		if err = json.Unmarshal(data, &loginData); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		var userData storage.UserData
		userData.Login = loginData.Login
		userData.Password = loginData.Password
		result, err := storage.CheckUserPassword(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !result {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		payload := jwt.MapClaims{
			"sub": loginData.Login,
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
	return http.HandlerFunc(login)
}
