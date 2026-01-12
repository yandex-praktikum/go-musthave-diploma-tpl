package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"gophermart/internal/service"
)

type BalanceHandler struct {
	balanceService *service.BalanceService
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := balanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case service.ErrInvalidOrderNumber:
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case service.ErrInsufficientFunds:
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	withdrawals, err := h.balanceService.ListWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type withdrawalResponse struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}

	response := make([]withdrawalResponse, 0, len(withdrawals))
	for _, wd := range withdrawals {
		response = append(response, withdrawalResponse{
			Order:       wd.Order,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
