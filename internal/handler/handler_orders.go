package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

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

	username := usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	err = h.svc.LoadOrder(ctx, string(orderNumber), username)
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
	username := usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	orders, err := h.svc.ListOrders(ctx, username)
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
