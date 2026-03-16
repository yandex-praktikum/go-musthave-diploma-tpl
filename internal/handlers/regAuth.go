package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/NailUsmanov/internal/storage"
	"github.com/NailUsmanov/models"
	"go.uber.org/zap"
)

func Register(s storage.Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sugar.Infof(">>> Register endpoint called")
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}
		// Декодим наш запрос
		var req models.RegistrationJSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sugar.Error("cannot decode request JSON body:", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		// Проверяем чтобы логин и пароль были не пустыми
		if len(req.Login) == 0 || len(req.Password) == 0 {
			http.Error(w, "empty login or password", http.StatusBadRequest)
			return
		}
		// Хэшируем пароль
		passwordHash := storage.HashPassword(req.Password)
		// Регистрируем пользователя
		err := s.Registration(r.Context(), req.Login, passwordHash)
		if err != nil {
			if errors.Is(err, storage.ErrOrderAlreadyUsed) {
				sugar.Errorf("Save error: %v", err)
				http.Error(w, "login is already occupied", http.StatusConflict)
				return
			}
			sugar.Errorf("Unexpected registration error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		// Получаем userID по login
		userID, err := s.GetUserIDByLogin(r.Context(), req.Login)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		// Возвращаем ответ
		sugar.Infof("User %s successfully registered", req.Login)
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    strconv.Itoa(userID),
			Path:     "/",
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
	}
}

func Login(s storage.Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}
		// Декодим запрос
		var req models.RegistrationJSON
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			sugar.Error("cannot decode request JSON body:", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		// Провеверяем чтобы логин и пароль не были пустыми
		if len(req.Login) == 0 || len(req.Password) == 0 {
			http.Error(w, "empty login or password", http.StatusBadRequest)
			return
		}
		// Проверяем наличие логина и совпадение хэша пароля в базе
		_, err = s.GetUserByLogin(r.Context(), req.Login)
		if err != nil {
			sugar.Errorf("Unexpected auth error: %v", err)
			http.Error(w, "can't find user", http.StatusUnauthorized)
			return
		}
		err = s.CheckHashMatch(r.Context(), req.Login, req.Password)
		if err != nil {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}
		// Получаем UserID по логину
		userID, err := s.GetUserIDByLogin(r.Context(), req.Login)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		// Устанавливаем куку и возвращаем ответ
		sugar.Infof("User %s successfully authenticated", req.Login)
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    strconv.Itoa(userID),
			Path:     "/",
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
	}
}
