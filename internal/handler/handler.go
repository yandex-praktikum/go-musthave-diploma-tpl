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

const queryTimeout = 3 * time.Second

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
	req.Body = http.MaxBytesReader(w, req.Body, maxRequestBodySize)

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

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
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

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
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

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

// loadOrder handles POST /api/user/orders.
func (h *apiHandler) loadOrder(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, maxRequestBodySize)
	orderNumber, err := io.ReadAll(req.Body)
	if maxBytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
		h.logger.Debug("content too large", zap.Error(maxBytesError))
		http.Error(w, "Content Too Large", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		h.logger.Debug("error reading body", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var user model.User
	user.Username = usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	err = h.svc.LoadOrder(ctx, string(orderNumber), user)
	if errors.Is(err, model.ErrMissingOrderNumber) || errors.Is(err, model.ErrInvalidOrderNumber) {
		h.logger.Debug("invalid order number", zap.Error(err))
		http.Error(w, "Unprocessable Content", http.StatusUnprocessableEntity)
		return
	}
	if errors.Is(err, model.ErrOrderExistsOtherUser) {
		h.logger.Debug("order already loaded by another user", zap.Error(err))
		http.Error(w, "Conflict", http.StatusConflict)
		return
	}
	if errors.Is(err, model.ErrOrderExistsSameUser) {
		h.logger.Debug("order already loaded by the same user", zap.Error(err))
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		h.logger.Error("failed to load order", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// listOrders handles GET /api/user/orders.
func (h *apiHandler) listOrders(w http.ResponseWriter, req *http.Request) {
	var user model.User

	user.Username = usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	orders, err := h.svc.ListOrders(ctx, user)
	if err != nil {
		h.logger.Error("failed to list orders", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}
	if orders == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		h.logger.Error("failed to marshal orders", zap.Error(err))
		return
	}
}

// getBalance handles GET /api/user/balance.
func (h *apiHandler) getBalance(w http.ResponseWriter, req *http.Request) {
	var user model.User

	user.Username = usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	balance, err := h.svc.GetBalance(ctx, user)
	if err != nil {
		h.logger.Error("failed to get balance", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		h.logger.Error("failed to marshal balance response", zap.Error(err))
		return
	}
}

// withdraw handles POST /api/user/balance/withdraw.
func (h *apiHandler) withdraw(w http.ResponseWriter, req *http.Request) {
	requestBody, ok := decodeJSON[model.WithdrawalRequest](w, req, h.logger)
	if !ok {
		return
	}

	username := usernameFromContext(req.Context())
	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	err := h.svc.Withdraw(ctx, username, requestBody)

	if errors.Is(err, model.ErrInsufficientFunds) {
		http.Error(w, "Payment Required", http.StatusPaymentRequired)
		return
	}
	if errors.Is(err, model.ErrInvalidOrderNumber) {
		http.Error(w, "Unprocessable Content", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		h.logger.Error("failed to withdraw", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// listWithdrawals handles GET /api/user/withdrawals.
func (h *apiHandler) listWithdrawals(w http.ResponseWriter, req *http.Request) {
	var user model.User

	user.Username = usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	withdrawals, err := h.svc.ListWithdrawals(ctx, user)
	if err != nil {
		h.logger.Error("failed to list withdrawals", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}
	if withdrawals == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(withdrawals)
	if err != nil {
		h.logger.Error("failed to marshal withdrawals", zap.Error(err))
		return
	}
}
