package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
)

type RegisterRequest struct {
	Name     string `json:"login"`
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
		userData.Login = regData.Name
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
		res.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(register)
}
