package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/gophermart"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"go.uber.org/zap"
)

func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userCredential dto.UserCredential
	err := decoder.Decode(&userCredential)
	if err != nil {
		logger.Error("Failed to decode UserCredential: %v", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.gophermart.RegisterUser(r.Context(), userCredential)
	if err != nil {
		logger.Error("Failed to register user: %v", zap.Error(err))
		w.WriteHeader(http.StatusConflict)
		return
	}
}

func (s *Server) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userCredential dto.UserCredential
	err := decoder.Decode(&userCredential)
	if err != nil {
		logger.Error("Failed to decode UserCredential", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userData, err := s.gophermart.LoginUser(r.Context(), userCredential)
	if err != nil {
		if errors.Is(err, gophermart.ErrUserNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
		}
		if errors.Is(err, gophermart.ErrIncorrectPassword) {
			w.WriteHeader(http.StatusUnauthorized)
		}
		return
	}

	s.setCookie(w, userData)
}

func (s *Server) setCookie(w http.ResponseWriter, userData *dto.UserData) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	sid := hex.EncodeToString(b)

	s.cookieHandler.Set(sid, userData)
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: sid,
	})
}
