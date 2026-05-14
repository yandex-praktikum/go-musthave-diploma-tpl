package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	apperror "github.com/AnxVit/go-musthave-diploma-tpl/internal/errors"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/utils"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Хэндлер регистрации пользователей
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	var req RequestRegister

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return apperror.NewBadRequestError("invalid request", "couldn't parse json")
	}

	ctx := r.Context()

	id, err := h.service.Register(ctx, &model.LoginPassword{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, model.ErrorUserAlreadyExists) {
			return apperror.NewConflictError("user already exists")
		}
	}

	r = r.WithContext(utils.SetUserID(ctx, id))

	h.setCookie(w, r)

	SendJSONResponse(w, "", http.StatusOK)
	return nil
}

// Хэндлер аутентификации пользователей
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	var req RequestLogin

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return apperror.NewBadRequestError("invalid request", "couldn't parse json")
	}

	ctx := r.Context()

	id, err := h.service.Login(ctx, &model.LoginPassword{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, model.ErrorNoUser) {
			return apperror.NewUnauthorizedError("user doesn't exists")
		}
		if errors.Is(err, model.ErrorUnverifiedAccount) {
			return apperror.NewUnauthorizedError("invalid password")
		}
		return apperror.NewInternalServerError(err.Error())
	}

	r = r.WithContext(utils.SetUserID(ctx, id))

	h.setCookie(w, r)

	SendJSONResponse(w, "", http.StatusOK)
	return nil
}

// Создание JWT токена и отправка его в cookie
func (h *Handler) setCookie(w http.ResponseWriter, r *http.Request) {
	expirationTime := time.Now().Add(time.Minute * time.Duration(h.cfg.ExpirationTimePerMinute))

	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(h.cfg.EncryptionKey)}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return
	}

	ctx := r.Context()

	userID := utils.GetUserID(ctx)

	cl := jwt.Claims{
		Subject:   userID,
		Expiry:    jwt.NewNumericDate(expirationTime),
		NotBefore: jwt.NewNumericDate(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	tokenString, err := jwt.Signed(sig).Claims(cl).CompactSerialize()
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
