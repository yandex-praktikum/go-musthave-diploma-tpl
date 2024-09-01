package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/eac0de/gophermart/internal/errors"
	"github.com/eac0de/gophermart/internal/services"
	"github.com/eac0de/gophermart/pkg/middlewares"
)

type BalanceHandlers struct {
	balanceService *services.BalanceService
}

func NewBalanceHandlers(balanceService *services.BalanceService) *BalanceHandlers {
	return &BalanceHandlers{
		balanceService: balanceService,
	}
}

func (bh *BalanceHandlers) BalanceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := middlewares.GetUserFromRequest(r)
	resBody := map[string]interface{}{
		"current":   user.Balance,
		"withdrawn": user.Withdrawn,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resBody)
}

func (bh *BalanceHandlers) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := middlewares.GetUserFromRequest(r)
	withdrawals, err := bh.balanceService.GetUserWithdrawals(r.Context(), user.ID)
	if err != nil {
		message, statusCode := errors.GetMessageAndStatusCode(err)
		http.Error(w, message, statusCode)
		return
	}
	if len(withdrawals) == 0 {
		http.Error(w, "withdrawals not found", http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)

}

func (bh *BalanceHandlers) CreateWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := middlewares.GetUserFromRequest(r)
	var reqBody struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	withdrawals, err := bh.balanceService.CreateWithdraw(r.Context(), reqBody.Order, reqBody.Sum, user.ID)
	if err != nil {
		message, statusCode := errors.GetMessageAndStatusCode(err)
		http.Error(w, message, statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
