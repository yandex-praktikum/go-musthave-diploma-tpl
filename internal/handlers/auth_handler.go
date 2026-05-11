// Package handlers предоставляет HTTP-хендлеры для API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-musthave-diploma-tpl/internal/auth"
	apperrors "go-musthave-diploma-tpl/internal/errors"
	"go-musthave-diploma-tpl/internal/logger"
	"go-musthave-diploma-tpl/internal/services"
)

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler создаёт новый хендлер аутентификации
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest запрос регистрации
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterResponse ответ регистрации/логина
type RegisterResponse struct {
	Token string `json:"token"`
}

// Register обрабатывает POST /api/user/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Debug("Failed to decode register request", "error", err)
		http.Error(w, apperrors.GetErrorMessage(apperrors.ErrInvalidRequest), http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	if req.Login == "" || req.Password == "" {
		logger.Debug("Empty login or password in register request")
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя
	token, err := h.userService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrLoginAlreadyExists) {
			logger.Info("Register failed: login already exists", "login", req.Login)
			http.Error(w, "Login already exists", http.StatusConflict)
			return
		}
		logger.ErrorWithStack("Register failed", err, "login", req.Login)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("User registered successfully", "login", req.Login)

	// Возвращаем токен
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+token)
	json.NewEncoder(w).Encode(RegisterResponse{Token: token})
}

// LoginRequest запрос логина
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Login обрабатывает POST /api/user/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Debug("Failed to decode login request", "error", err)
		http.Error(w, apperrors.GetErrorMessage(apperrors.ErrInvalidRequest), http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	if req.Login == "" || req.Password == "" {
		logger.Debug("Empty login or password in login request")
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Аутентифицируем пользователя
	token, err := h.userService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			logger.Info("Login failed: invalid credentials", "login", req.Login)
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}
		logger.ErrorWithStack("Login failed", err, "login", req.Login)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("User logged in successfully", "login", req.Login)

	// Возвращаем токен
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+token)
	json.NewEncoder(w).Encode(RegisterResponse{Token: token})
}

// GetUserID извлекает userID из контекста
func GetUserID(r *http.Request) (int64, bool) {
	return auth.GetUserID(r.Context())
}
