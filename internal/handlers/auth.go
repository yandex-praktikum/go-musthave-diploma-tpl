// internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/b602op/go-musthave-diploma-tpl/internal/middleware"
	"github.com/b602op/go-musthave-diploma-tpl/internal/service"
)

// AuthHandler — HTTP-обработчики регистрации и аутентификации.
type AuthHandler struct {
	authService    *service.AuthService
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthHandler создаёт новый AuthHandler с указанным сервисом
// аутентификации и middleware авторизации.
func NewAuthHandler(
	authService *service.AuthService,
	authMiddleware *middleware.AuthMiddleware,
) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		authMiddleware: authMiddleware,
	}
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrLoginAlreadyExists) {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем токен через authMiddleware
	token, err := h.authMiddleware.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
	})

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login обрабатывает аутентификацию пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем токен через authMiddleware
	token, err := h.authMiddleware.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
	})

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
