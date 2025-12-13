package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/auth"
	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/models"
	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/storage/postgres"
	"golang.org/x/crypto/bcrypt"
)

// checkConnectDatabase проверяет подключение к базе данных
func checkConnectDatabase(dbStorage postgres.DatabaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if dbStorage == nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Database not configured - using in-memory storage"))
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		if err := dbStorage.Ping(ctx); err != nil {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Database connection successful"))
	}
}

// registrationUsers обрабатывает регистрацию пользователя
func registrationUsers(dbStorage postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}

		// Декодируем тело запроса
		var req models.AuthRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Валидация входных данных
		if req.Login == "" || req.Password == "" {
			http.Error(w, "Login and password are required", http.StatusBadRequest)
			return
		}

		// Проверяем существование пользователя
		ctx := r.Context()
		exists, err := dbStorage.(interface {
			UserExists(ctx context.Context, login string) (bool, error)
		}).UserExists(ctx, req.Login)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if exists {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Создаем пользователя
		err = dbStorage.(interface {
			CreateUser(ctx context.Context, login, hashedPassword string) error
		}).CreateUser(ctx, req.Login, string(hashedPassword))

		if err != nil {
			// Проверяем ошибку на конфликт
			if err.Error() == fmt.Sprintf("user with login '%s' already exists", req.Login) {
				http.Error(w, "Login already taken", http.StatusConflict)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Получаем созданного пользователя
		user, err := dbStorage.(interface {
			GetUserByLogin(ctx context.Context, login string) (models.User, error)
		}).GetUserByLogin(ctx, req.Login)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Создаем JWT токен
		token, err := auth.CreateToken(user.ID, user.Login)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем токен в cookie
		middleware.SetAuthToken(w, token)

		// Отправляем успешный ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User successfully registered and authenticated",
			"user_id": user.ID,
		})
	}
}

// authUsers обрабатывает аутентификацию пользователя
func authUsers(dbStorage postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}

		// Декодируем тело запроса
		var req models.AuthRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Валидация входных данных
		if req.Login == "" || req.Password == "" {
			http.Error(w, "Login and password are required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Получаем пользователя
		user, err := dbStorage.(interface {
			GetUserByLogin(ctx context.Context, login string) (models.User, error)
		}).GetUserByLogin(ctx, req.Login)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Проверяем пароль
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		// Создаем JWT токен
		token, err := auth.CreateToken(user.ID, user.Login)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем токен в cookie
		middleware.SetAuthToken(w, token)

		// Отправляем успешный ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User successfully authenticated",
			"user_id": user.ID,
		})
	}
}
