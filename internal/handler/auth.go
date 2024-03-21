package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/SmoothWay/gophermart/internal/logger"
	"github.com/SmoothWay/gophermart/internal/model"
	"github.com/SmoothWay/gophermart/internal/service/auth"

	"go.uber.org/zap"
)

const backendTimeout = 3 * time.Second

type authenticationHandler struct {
	authenticationService auth.AuthenticationService
}

func New(authenticationService auth.AuthenticationService) authenticationHandler {
	return authenticationHandler{
		authenticationService: authenticationService,
	}
}

// signup user
func (h *authenticationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var reqBody model.UserRegisterRequest

	err := readJSON(w, r, &reqBody)
	if err != nil {
		logger.Log().Info("req", zap.Error(err))
		errorJSON(w, errors.New("bad request"), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), backendTimeout)
	defer cancel()

	token, err := h.authenticationService.Register(ctx, reqBody.Login, string(reqBody.Password))
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExist) {
			logger.Log().Info("Register User", zap.Any(reqBody.Login, err))
			errorJSON(w, auth.ErrUserAlreadyExist, http.StatusConflict)
			return
		}
		logger.Log().Info("Register User", zap.Any(reqBody.Login, err))
		errorJSON(w, ErrInternalError, http.StatusInternalServerError)
		return
	}

	response := model.UserRegisterResponse{
		Token:  token,
		Status: "Created",
	}

	writeJSON(w, http.StatusCreated, response)
}

// login user
func (h *authenticationHandler) Login(w http.ResponseWriter, r *http.Request) {
	var reqBody model.UserRegisterRequest
	err := readJSON(w, r, reqBody)
	if err != nil {
		logger.Log().Info("req", zap.Error(err))
		errorJSON(w, errors.New("bad request"), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), backendTimeout)
	defer cancel()

	token, err := h.authenticationService.Login(ctx, reqBody.Login, string(reqBody.Password))
	if err != nil {
		if errors.Is(err, auth.ErrIncorrectLoginOrPassword) {
			logger.Log().Info("Register GetUserByEmail", zap.Any(reqBody.Login, err))
			errorJSON(w, auth.ErrIncorrectLoginOrPassword, http.StatusConflict)
			return
		}
		logger.Log().Info("Register GetUserByEamil", zap.Any(reqBody.Login, err))
		errorJSON(w, ErrInternalError, http.StatusInternalServerError)
		return
	}

	response := model.UserLoginResponse{
		Token: token,
	}

	writeJSON(w, http.StatusOK, response)
}
