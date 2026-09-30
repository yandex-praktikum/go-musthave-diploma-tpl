package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
	"github.com/nikitaw13/go-musthave-diploma/internal/service"
	"go.uber.org/zap"
)

// apiHandler serves the gophermart HTTP API on top of a LoyaltyService.
type apiHandler struct {
	svc    *service.LoyaltyService
	logger *zap.Logger
}

// NewAPIHandler creates a new API handler for the given loyalty service and logger.
func NewAPIHandler(svc *service.LoyaltyService, logger *zap.Logger) *apiHandler {
	return &apiHandler{
		svc:    svc,
		logger: logger,
	}
}

// decodeJSON reads the request body into T and writes an HTTP error response on failure;
// the bool reports success.
func decodeJSON[T any](w http.ResponseWriter, req *http.Request, logger *zap.Logger) (T, bool) {
	var v T
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if maxBytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
		logger.Debug("content too large", zap.Error(maxBytesError))
		http.Error(w, "Content Too Large", http.StatusRequestEntityTooLarge)
		return v, false
	}
	if err != nil {
		logger.Error("error reading buffer", zap.Error(err))
		http.Error(w, "The request contains bad syntax or cannot be fulfilled", http.StatusBadRequest)
		return v, false
	}

	err = json.Unmarshal(buf.Bytes(), &v)
	if syntaxErr, ok := errors.AsType[*json.SyntaxError](err); ok {
		logger.Error("invalid JSON syntax", zap.Error(err))
		http.Error(w, fmt.Sprintf("invalid JSON syntax at offset %d", syntaxErr.Offset), http.StatusBadRequest)
		return v, false
	}
	if unmarshalTypeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		logger.Error("invalid type for field", zap.Error(err))
		http.Error(w, fmt.Sprintf("invalid type %q for field %q", unmarshalTypeErr.Value, unmarshalTypeErr.Field), http.StatusBadRequest)
		return v, false
	}

	if err != nil {
		logger.Error("error unmarshaling", zap.Error(err))
		http.Error(w, "The request contains bad syntax or cannot be fulfilled", http.StatusBadRequest)
		return v, false
	}
	return v, true
}

// registerUser handles POST /api/user/register.
func (h *apiHandler) registerUser(w http.ResponseWriter, req *http.Request) {
	user, ok := decodeJSON[model.User](w, req, h.logger)
	if !ok {
		return
	}

	err := user.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	token, err := h.svc.RegisterUser(ctx, user)
	if errors.Is(err, model.ErrUserAlreadyExists) {
		h.logger.Debug("user already exists", zap.String("username", user.Username))
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	if err != nil {
		h.logger.Error("failed to register user", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	if token == "" {
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)

}

// loginUser handles POST /api/user/login.
func (h *apiHandler) loginUser(w http.ResponseWriter, req *http.Request) {
	user, ok := decodeJSON[model.User](w, req, h.logger)
	if !ok {
		return
	}

	err := user.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	token, err := h.svc.LoginUser(ctx, user)
	if errors.Is(err, model.ErrUserNotFound) || errors.Is(err, model.ErrBadPassword) || errors.Is(err, model.ErrMismatchedPassword) {
		h.logger.Debug(err.Error(), zap.String("username", user.Username))
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		h.logger.Error("failed to login user", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

// 200 — номер заказа уже был загружен этим пользователем;
// 202 — новый номер заказа принят в обработку;
// 400 — неверный формат запроса;									++
// 401 — пользователь не аутентифицирован;							+++реализовано в middleware
// 409 — номер заказа уже был загружен другим пользователем;
// 422 — неверный формат номера заказа;								+++реализовано в model
// 500 — внутренняя ошибка сервера.

// TODO: implement
func (h *apiHandler) loadOrder(w http.ResponseWriter, req *http.Request) {
	orderNumber, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Debug("error reading body", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var user model.User
	user.Username = usernameFromContext(req.Context())

	err = h.svc.LoadOrder(req.Context(), string(orderNumber), user)
	if errors.Is(err, model.ErrMissingOrderNumber) {
		h.logger.Debug("error ErrMissingOrderNumber", zap.Error(err))
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if errors.Is(err, model.ErrInvalidOrderNumber) {
		h.logger.Debug("error ErrInvalidOrderNumber", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		h.logger.Debug("error StatusBadRequest", zap.Error(err))
		http.Error(w, "The request contains bad syntax or cannot be fulfilled", http.StatusBadRequest)
		return
	}

}

// TODO: implement
func (h *apiHandler) listOrders(w http.ResponseWriter, req *http.Request) {
}

// TODO: implement
func (h *apiHandler) getBalance(w http.ResponseWriter, req *http.Request) {
}

// TODO: implement
func (h *apiHandler) withdraw(w http.ResponseWriter, req *http.Request) {
}

// TODO: implement
func (h *apiHandler) listWithdrawals(w http.ResponseWriter, req *http.Request) {
}
