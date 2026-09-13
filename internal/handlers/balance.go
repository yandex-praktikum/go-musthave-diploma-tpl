package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/b602op/go-musthave-diploma-tpl/internal/middleware"
	"github.com/b602op/go-musthave-diploma-tpl/internal/service"
)

// BalanceHandler — HTTP-обработчики баланса и списаний.
type BalanceHandler struct {
	balanceService *service.BalanceService
}

// NewBalanceHandler создаёт новый BalanceHandler с указанным сервисом баланса.
func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

// GetBalance возвращает текущий баланс пользователя
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(userID)
	if err != nil {
		log.Printf("GetBalance error for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := domain.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("GetBalance JSON encode error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Withdraw обрабатывает запрос на списание баллов
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Withdraw decode error: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Printf("Withdraw: userID=%d, order=%s, sum=%f", userID, req.Order, req.Sum)

	err := h.balanceService.Withdraw(userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case domain.ErrInvalidOrderNumber:
			log.Printf("Withdraw: invalid order number: %s", req.Order)
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		case domain.ErrInsufficientBalance:
			log.Printf("Withdraw: insufficient balance for user %d", userID)
			http.Error(w, "Insufficient balance", http.StatusPaymentRequired)
		default:
			log.Printf("Withdraw error for user %d: %v", userID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Withdraw: success for user %d, order %s", userID, req.Order)
	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals возвращает историю списаний
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	log.Println("=== GetWithdrawals called ===")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		log.Println("GetWithdrawals: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("GetWithdrawals: userID=%d", userID)

	withdrawals, err := h.balanceService.GetWithdrawals(userID)
	if err != nil {
		log.Printf("GetWithdrawals: service error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("GetWithdrawals: found %d withdrawals", len(withdrawals))

	if len(withdrawals) == 0 {
		log.Println("GetWithdrawals: no withdrawals, returning 204")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Конвертируем в response DTO
	response := make([]domain.WithdrawalResponse, len(withdrawals))
	for i, wd := range withdrawals {
		response[i] = domain.WithdrawalResponse{
			Order:       wd.OrderNumber,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("GetWithdrawals: JSON encode error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("GetWithdrawals: success")
}
