package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

// withdraw handles POST /api/user/balance/withdraw.
func (h *apiHandler) withdraw(w http.ResponseWriter, req *http.Request) {
	withdrawalRequest, ok := decodeJSON[model.WithdrawalRequest](w, req, h.logger)
	if !ok {
		return
	}

	username := usernameFromContext(req.Context())
	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	err := h.svc.Withdraw(ctx, username, withdrawalRequest)
	if errors.Is(err, model.ErrNonPositiveAmount) {
		h.logger.Debug("the withdrawal amount is zero or negative", zap.Error(err))
		http.Error(w, "Unprocessable Content", http.StatusUnprocessableEntity)
		return
	}

	if errors.Is(err, model.ErrInsufficientFunds) {
		http.Error(w, "Payment Required", http.StatusPaymentRequired)
		return
	}
	if errors.Is(err, model.ErrMissingOrderNumber) || errors.Is(err, model.ErrInvalidOrderNumber) {
		h.logger.Debug("invalid order number", zap.Error(err))
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
	username := usernameFromContext(req.Context())

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	withdrawals, err := h.svc.ListWithdrawals(ctx, username)
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
