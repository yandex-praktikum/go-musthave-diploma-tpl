package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eac0de/gophermart/internal/services"
	"github.com/eac0de/gophermart/pkg/jwt"
	"github.com/eac0de/gophermart/pkg/middlewares"
)

type AuthHandlers struct {
	tokenService *jwt.JWTTokenService
	authService  *services.AuthService
}

func NewAuthHandlers(tokenService *jwt.JWTTokenService, authService *services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		tokenService: tokenService,
		authService:  authService,
	}
}

func (ah *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	user, err := ah.authService.CreateUser(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokens, err := ah.tokenService.GenerateTokens(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	authorizationHeader := fmt.Sprintf("Bearer %s", tokens.AccessToken)
	w.Header().Set("Authorization", authorizationHeader)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}

func (ah *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	user, err := ah.authService.GetUser(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokens, err := ah.tokenService.GenerateTokens(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	authorizationHeader := fmt.Sprintf("Bearer %s", tokens.AccessToken)
	w.Header().Set("Authorization", authorizationHeader)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}

func (ah *AuthHandlers) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&requestBody)
	if requestBody.RefreshToken == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	claims, err := ah.tokenService.ValidateRefreshToken(r.Context(), requestBody.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	newAccessToken, err := ah.tokenService.BuildAccessToken(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}
	responseBody := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: newAccessToken,
	}
	authorizationHeader := fmt.Sprintf("Bearer %s", newAccessToken)
	w.Header().Set("Authorization", authorizationHeader)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func (ah *AuthHandlers) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromRequest(r)
	var requestBody struct {
		Password *string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if requestBody.Password == nil {
		http.Error(w, "Invalid request payload ", http.StatusBadRequest)
		return
	}
	err := ah.authService.ChangePassword(r.Context(), user.Username, *requestBody.Password)
	if err != nil {
		http.Error(w, "Failed to update password the account", http.StatusBadRequest)
		return
	}
	tokens, err := ah.tokenService.GenerateTokens(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}
