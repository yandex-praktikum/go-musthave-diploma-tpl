package handler

import (
	"encoding/json"
	"net/http"

	"gophermart/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtService  *service.JWTService
}

func NewAuthHandler(authService *service.AuthService, jwtService *service.JWTService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
	}
}

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var cred credentials
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := h.authService.Register(r.Context(), cred.Login, cred.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case service.ErrConflict:
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	token, err := h.jwtService.GenerateToken(userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setJWTCookie(w, token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var cred credentials
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := h.authService.Login(r.Context(), cred.Login, cred.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case service.ErrUnauthorized:
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	token, err := h.jwtService.GenerateToken(userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setJWTCookie(w, token)
	w.WriteHeader(http.StatusOK)
}

func setJWTCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}
