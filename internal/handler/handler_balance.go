package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// getBalance handles GET /api/user/balance.
func (h *apiHandler) getBalance(w http.ResponseWriter, req *http.Request) {
	username := usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	balance, err := h.svc.GetBalance(ctx, username)
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
