package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	apperrors "go-musthave-diploma-tpl/internal/errors"
	"go-musthave-diploma-tpl/internal/logger"
	"go-musthave-diploma-tpl/internal/services"
	"go-musthave-diploma-tpl/internal/storage"
)

// OrderHandler обрабатывает запросы к заказам
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler создаёт новый хендлер заказов
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// OrderResponse ответ для заказа
type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// UploadOrder обрабатывает POST /api/user/orders
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := GetUserID(r)
	if !ok {
		logger.Warn("UploadOrder: unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем тело запроса (text/plain)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Debug("Failed to read order request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	number := string(body)

	logger.Debug("Uploading order", "user_id", userID, "number", number)

	// Загружаем заказ
	exists, conflict, err := h.orderService.UploadOrder(r.Context(), userID, number)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOrderNumber) {
			logger.Info("UploadOrder: invalid order number", "number", number)
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}
		logger.ErrorWithStack("UploadOrder: failed to upload order", err, "number", number)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		logger.Debug("UploadOrder: order already exists for this user", "number", number)
		w.WriteHeader(http.StatusOK)
		return
	}

	if conflict {
		logger.Info("UploadOrder: order belongs to another user", "number", number)
		http.Error(w, "Order number already exists", http.StatusConflict)
		return
	}

	logger.Info("Order uploaded successfully", "user_id", userID, "number", number)
	w.WriteHeader(http.StatusAccepted)
}

// GetOrders обрабатывает GET /api/user/orders
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := GetUserID(r)
	if !ok {
		logger.Warn("GetOrders: unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем заказы
	orders, err := h.orderService.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		logger.ErrorWithStack("GetOrders: failed to get orders", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		logger.Debug("GetOrders: no orders found", "user_id", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Формируем ответ
	response := make([]OrderResponse, len(orders))
	for i, order := range orders {
		response[i] = OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
	}

	logger.Debug("GetOrders: returning orders", "user_id", userID, "count", len(orders))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BalanceResponse ответ для баланса
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// GetBalance обрабатывает GET /api/user/balance
func (h *OrderHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := GetUserID(r)
	if !ok {
		logger.Warn("GetBalance: unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем баланс
	balance, err := h.orderService.GetBalance(r.Context(), userID)
	if err != nil {
		logger.ErrorWithStack("GetBalance: failed to get balance", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Debug("GetBalance: returning balance", "user_id", userID, "current", balance.Current)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}

// WithdrawRequest запрос на списание
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Withdraw обрабатывает POST /api/user/balance/withdraw
func (h *OrderHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := GetUserID(r)
	if !ok {
		logger.Warn("Withdraw: unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Debug("Withdraw: failed to decode request", "error", err)
		http.Error(w, apperrors.GetErrorMessage(apperrors.ErrInvalidRequest), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Order == "" || req.Sum <= 0 {
		logger.Debug("Withdraw: invalid request body", "order", req.Order, "sum", req.Sum)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Debug("Withdraw request", "user_id", userID, "order", req.Order, "sum", req.Sum)

	// Списываем средства
	err := h.orderService.Withdraw(r.Context(), userID, services.WithdrawRequest{
		Order: req.Order,
		Sum:   req.Sum,
	})
	if err != nil {
		if errors.Is(err, services.ErrInvalidOrderNumber) {
			logger.Info("Withdraw: invalid order number", "order", req.Order)
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, storage.ErrInsufficientFunds) {
			logger.Info("Withdraw: insufficient funds", "user_id", userID, "required", req.Sum)
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		logger.ErrorWithStack("Withdraw: failed to withdraw", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("Withdraw successful", "user_id", userID, "order", req.Order, "sum", req.Sum)
	w.WriteHeader(http.StatusOK)
}

// WithdrawalResponse ответ для списания
type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// GetWithdrawals обрабатывает GET /api/user/withdrawals
func (h *OrderHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := GetUserID(r)
	if !ok {
		logger.Warn("GetWithdrawals: unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем списания
	withdrawals, err := h.orderService.GetWithdrawalsByUser(r.Context(), userID)
	if err != nil {
		logger.ErrorWithStack("GetWithdrawals: failed to get withdrawals", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		logger.Debug("GetWithdrawals: no withdrawals found", "user_id", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Формируем ответ
	response := make([]WithdrawalResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		response[i] = WithdrawalResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		}
	}

	logger.Debug("GetWithdrawals: returning withdrawals", "user_id", userID, "count", len(withdrawals))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
