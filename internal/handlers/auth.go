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
	TokenService *jwt.JWTTokenService
	AuthService  *services.AuthService
}

func NewAuthHandlers(TokenService *jwt.JWTTokenService, AuthService *services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		TokenService: TokenService,
		AuthService:  AuthService,
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
	user, err := ah.AuthService.CreateUser(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokens, err := ah.TokenService.GenerateTokens(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	authorizationHeader := fmt.Sprintf("Bearer %s", tokens.AccessToken)
	w.Header().Set("Authorization", authorizationHeader)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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
	user, err := ah.AuthService.GetUser(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokens, err := ah.TokenService.GenerateTokens(r.Context(), user.ID)
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
	claims, err := ah.TokenService.ValidateRefreshToken(r.Context(), requestBody.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	newAccessToken, err := ah.TokenService.BuildAccessToken(r.Context(), claims.UserID)
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
	err := ah.AuthService.ChangePassword(r.Context(), user.Username, *requestBody.Password)
	if err != nil {
		http.Error(w, "Failed to update password the account", http.StatusBadRequest)
		return
	}
	tokens, err := ah.TokenService.GenerateTokens(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}
